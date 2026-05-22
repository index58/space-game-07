package world

import (
	"errors"
	"math"
	"sort"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

const exchangeTaskAcronym = "Exchange"

type exchangeRequest struct {
	SenderCosmicObjectID   int64   // Объект, который отправил запрос.
	ReceiverCosmicObjectID int64   // Объект, который должен принять запрос.
	RemainingSeconds       float64 // Остаток времени ожидания ответа.
	DockedForExchange      bool    // Нужно ли отстыковать объекты после завершения.
	Approved               bool    // Принят ли запрос, ожидающий окончания стыковки.
}

type exchangeSession struct {
	SenderCosmicObjectID   int64                  // Объект отправителя запроса.
	ReceiverCosmicObjectID int64                  // Объект получателя запроса.
	DockedForExchange      bool                   // Нужно ли отстыковать объекты после завершения.
	Moving                 bool                   // Началось ли необратимое перемещение.
	Participants           [2]exchangeParticipant // Состояния двух сторон обмена.
}

type exchangeParticipant struct {
	ObjectID            int64                      // Объект участника.
	ReceiverContainerID int64                      // Контейнер для получаемых предметов.
	SourceContainerID   int64                      // Контейнер, из которого предметы кладутся в очередь.
	TaskID              int64                      // Задание, в котором лежит очередь отдаваемых предметов.
	Confirmed           bool                       // Подтвердил ли участник обмен.
	NotEnoughSpace      bool                       // Нужно ли показать участнику нехватку места.
	ProgressByGroupID   map[int64]exchangeProgress // Прогресс строк очереди.
}

type exchangeProgress struct {
	RemainingEnergy float64 // Остаток работы по строке.
	TotalEnergy     float64 // Полная работа по строке.
}

// SendExchangeRequest запускает запрос обмена от текущего управляемого объекта.
func (world *World) SendExchangeRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if world.exchangeObjectIsBusyLocked(sender.ID) {
		return errors.New("object already participates in exchange")
	}
	receiver, dockedForExchange, err := world.exchangeReceiverLocked(sender)
	if err != nil {
		return err
	}
	if receiver.OwnerCharacterID == sender.OwnerCharacterID {
		return errors.New("exchange requires another player")
	}
	if !world.dockingReceiverHasDecisionMakerLocked(receiver.ID) {
		world.addExchangeNotificationLocked([]int64{sender.ID}, "В объекте назначения нет персонажа для принятия обмена")
		return nil
	}
	world.exchangeRequests = append(world.exchangeRequests, exchangeRequest{
		SenderCosmicObjectID:   sender.ID,
		ReceiverCosmicObjectID: receiver.ID,
		RemainingSeconds:       dockingDurationSeconds,
		DockedForExchange:      dockedForExchange,
	})
	world.addExchangeWindowEventsLocked("exchangeRequestStarted", sender.ID, receiver.ID, dockingDurationSeconds)
	return nil
}

// ApproveExchangeRequest принимает входящий запрос обмена.
func (world *World) ApproveExchangeRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	index := world.exchangeRequestIndexByReceiverLocked(receiver.ID)
	if index < 0 {
		return errors.New("exchange request not found")
	}
	request := world.exchangeRequests[index]
	sender, ok := world.data.CosmicObjects.Get(request.SenderCosmicObjectID)
	if !ok {
		return errors.New("sender object not found")
	}
	if request.DockedForExchange {
		request.Approved = true
		world.exchangeRequests[index] = request
		world.startDockingProcessLocked(sender.ID, receiver.ID)
		return nil
	}
	world.exchangeRequests = append(world.exchangeRequests[:index], world.exchangeRequests[index+1:]...)
	world.openExchangeSessionLocked(sender.ID, receiver.ID, request.DockedForExchange)
	return nil
}

// RejectExchangeRequest отклоняет входящий запрос обмена.
func (world *World) RejectExchangeRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	index := world.exchangeRequestIndexByReceiverLocked(receiver.ID)
	if index < 0 {
		return errors.New("exchange request not found")
	}
	request := world.exchangeRequests[index]
	world.exchangeRequests = append(world.exchangeRequests[:index], world.exchangeRequests[index+1:]...)
	world.addExchangeWindowEventsLocked("exchangeRejected", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID, 0)
	return nil
}

// SelectExchangeReceiver выбирает контейнер для получаемых предметов.
func (world *World) SelectExchangeReceiver(accountID int64, containerID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	session, participant, _, err := world.exchangeParticipantByAccountLocked(accountID)
	if err != nil {
		return err
	}
	if participant.Confirmed {
		return errors.New("receiver container is locked after confirmation")
	}
	if _, err := world.exchangeOwnedContainerLocked(participant.ObjectID, containerID); err != nil {
		return err
	}
	participant.ReceiverContainerID = containerID
	participant.NotEnoughSpace = false
	world.broadcastExchangeStateLocked(session)
	return nil
}

// SelectExchangeSource выбирает контейнер, из которого предметы кладутся в очередь.
func (world *World) SelectExchangeSource(accountID int64, containerID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	session, participant, peer, err := world.exchangeParticipantByAccountLocked(accountID)
	if err != nil {
		return err
	}
	if peer.Confirmed {
		return errors.New("queue is locked by other player confirmation")
	}
	container, err := world.exchangeOwnedContainerLocked(participant.ObjectID, containerID)
	if err != nil {
		return err
	}
	participant.SourceContainerID = container.ID
	if participant.TaskID > 0 {
		if task, ok := world.data.Tasks.Get(participant.TaskID); ok {
			task.ControllerEquipmentGroupID = container.ID
			_ = world.data.Tasks.RebuildIndexes()
		}
	}
	world.broadcastExchangeStateLocked(session)
	return nil
}

// AddExchangeItems кладет выбранные предметы в очередь обмена.
func (world *World) AddExchangeItems(accountID int64, itemGroupIDs []int64, amount float64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	session, participant, peer, err := world.exchangeParticipantByAccountLocked(accountID)
	if err != nil {
		return err
	}
	if peer.Confirmed {
		return errors.New("queue is locked by other player confirmation")
	}
	if participant.SourceContainerID <= 0 {
		return errors.New("source container is not selected")
	}
	if amount <= physics.Epsilon {
		return errors.New("exchange amount is empty")
	}
	task, err := world.exchangeTaskLocked(participant)
	if err != nil {
		return err
	}
	for _, itemGroupID := range itemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != participant.SourceContainerID {
			return errors.New("item group does not belong to source container")
		}
		count := itemGroup.Count
		if len(itemGroupIDs) == 1 {
			count = math.Min(itemGroup.Count, amount)
		}
		world.consumeItemModelFromContainerLocked(participant.SourceContainerID, itemGroup.ContentItemModelID, count)
		taskGroup, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{
			TaskID:      task.ID,
			ItemModelID: itemGroup.ContentItemModelID,
			Count:       count,
			IsStored:    true,
		})
		if err != nil {
			return err
		}
		participant.ProgressByGroupID[taskGroup.ID] = exchangeProgress{}
	}
	world.broadcastExchangeStateLocked(session)
	return nil
}

// ConfirmExchange проверяет приемник и подтверждает сторону обмена.
func (world *World) ConfirmExchange(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	session, participant, peer, err := world.exchangeParticipantByAccountLocked(accountID)
	if err != nil {
		return err
	}
	if participant.ReceiverContainerID <= 0 {
		return errors.New("receiver container is not selected")
	}
	peerGroups := world.exchangeTaskGroupsLocked(peer)
	if !world.containerCanAcceptTaskReserveLocked(participant.ReceiverContainerID, peerGroups) && len(peerGroups) > 0 {
		participant.NotEnoughSpace = true
		world.addExchangeNotificationLocked([]int64{participant.ObjectID}, "Недостаточно места")
		world.broadcastExchangeStateLocked(session)
		return nil
	}
	participant.Confirmed = true
	participant.NotEnoughSpace = false
	if session.Participants[0].Confirmed && session.Participants[1].Confirmed {
		session.Moving = true
		world.prepareExchangeProgressLocked(session)
	}
	world.broadcastExchangeStateLocked(session)
	return nil
}

// CancelExchange отменяет обмен до начала необратимого перемещения.
func (world *World) CancelExchange(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	index, session, err := world.exchangeSessionByAccountLocked(accountID)
	if err != nil {
		return err
	}
	if session.Moving {
		return errors.New("exchange is already moving")
	}
	world.cancelExchangeSessionLocked(session)
	world.exchangeSessions = append(world.exchangeSessions[:index], world.exchangeSessions[index+1:]...)
	world.addExchangeWindowEventsLocked("exchangeCancelled", session.SenderCosmicObjectID, session.ReceiverCosmicObjectID, 0)
	return nil
}

// DrainExchangeEvents забирает накопленные события обмена для сетевой рассылки.
func (world *World) DrainExchangeEvents() []game.ExchangeEvent {
	world.mu.Lock()
	defer world.mu.Unlock()

	events := append([]game.ExchangeEvent(nil), world.exchangeEvents...)
	world.exchangeEvents = world.exchangeEvents[:0]
	return events
}

// ExchangeEventsForAccount возвращает текущий обмен для повторно подключившегося клиента.
func (world *World) ExchangeEventsForAccount(accountID int64) []game.ExchangeEvent {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return nil
	}
	for _, request := range world.exchangeRequests {
		if request.SenderCosmicObjectID == objectID {
			return []game.ExchangeEvent{{
				Type:     "exchangeEvent",
				Kind:     "exchangeRequestStarted",
				Role:     "sender",
				Duration: request.RemainingSeconds,
			}}
		}
		if request.ReceiverCosmicObjectID == objectID {
			return []game.ExchangeEvent{{
				Type:     "exchangeEvent",
				Kind:     "exchangeRequestStarted",
				Role:     "receiver",
				Duration: request.RemainingSeconds,
			}}
		}
	}
	for index := range world.exchangeSessions {
		session := &world.exchangeSessions[index]
		if session.Participants[0].ObjectID == objectID {
			return []game.ExchangeEvent{{
				Type:  "exchangeEvent",
				Kind:  "exchangeState",
				Role:  "sender",
				State: world.exchangeStateForLocked(session, 0),
			}}
		}
		if session.Participants[1].ObjectID == objectID {
			return []game.ExchangeEvent{{
				Type:  "exchangeEvent",
				Kind:  "exchangeState",
				Role:  "receiver",
				State: world.exchangeStateForLocked(session, 1),
			}}
		}
	}
	return nil
}

func (world *World) stepExchangeSessionsLocked(dtSeconds float64) {
	if dtSeconds <= 0 {
		return
	}
	remaining := world.exchangeSessions[:0]
	for index := range world.exchangeSessions {
		session := &world.exchangeSessions[index]
		if !session.Moving {
			remaining = append(remaining, *session)
			continue
		}
		world.stepExchangeMovementLocked(session, dtSeconds)
		if world.exchangeSessionReadyLocked(session) {
			world.finishExchangeSessionLocked(session)
			world.addExchangeWindowEventsLocked("exchangeFinished", session.SenderCosmicObjectID, session.ReceiverCosmicObjectID, 0)
			continue
		}
		world.broadcastExchangeStateLocked(session)
		remaining = append(remaining, *session)
	}
	world.exchangeSessions = remaining
}

func (world *World) stepExchangeRequestsLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.exchangeRequests) == 0 {
		return
	}
	remaining := world.exchangeRequests[:0]
	for _, request := range world.exchangeRequests {
		if request.Approved {
			remaining = append(remaining, request)
			continue
		}
		request.RemainingSeconds -= dtSeconds
		if request.RemainingSeconds > physics.Epsilon {
			remaining = append(remaining, request)
			continue
		}
		world.addExchangeWindowEventsLocked("exchangeRejected", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID, 0)
	}
	world.exchangeRequests = remaining
}

func (world *World) exchangeReceiverLocked(sender *data.CosmicObject) (*data.CosmicObject, bool, error) {
	if sender.ClusterMainCosmicObjectID > 0 {
		targetID, ok := world.autoLandingTargetIDLocked(sender)
		if !ok {
			return nil, false, errors.New("exchange target not found")
		}
		receiver, ok := world.data.CosmicObjects.Get(targetID)
		if !ok {
			return nil, false, errors.New("exchange target not found")
		}
		return receiver, false, nil
	}
	receiver, err := world.findDockingReceiverLocked(sender)
	if err != nil {
		return nil, false, err
	}
	return receiver, true, nil
}

func (world *World) exchangeRequestIndexByReceiverLocked(receiverID int64) int {
	for index, request := range world.exchangeRequests {
		if request.ReceiverCosmicObjectID == receiverID && !request.Approved {
			return index
		}
	}
	return -1
}

func (world *World) openExchangeAfterDockingLocked(senderID int64, receiverID int64) {
	for index, request := range world.exchangeRequests {
		if !request.Approved || !request.DockedForExchange {
			continue
		}
		if request.SenderCosmicObjectID != senderID || request.ReceiverCosmicObjectID != receiverID {
			continue
		}
		world.exchangeRequests = append(world.exchangeRequests[:index], world.exchangeRequests[index+1:]...)
		world.openExchangeSessionLocked(senderID, receiverID, true)
		return
	}
}

func (world *World) openExchangeSessionLocked(senderID int64, receiverID int64, dockedForExchange bool) {
	session := exchangeSession{
		SenderCosmicObjectID:   senderID,
		ReceiverCosmicObjectID: receiverID,
		DockedForExchange:      dockedForExchange,
		Participants: [2]exchangeParticipant{
			{
				ObjectID:            senderID,
				ReceiverContainerID: world.defaultExchangeContainerLocked(senderID),
				SourceContainerID:   world.defaultExchangeContainerLocked(senderID),
				ProgressByGroupID:   map[int64]exchangeProgress{},
			},
			{
				ObjectID:            receiverID,
				ReceiverContainerID: world.defaultExchangeContainerLocked(receiverID),
				SourceContainerID:   world.defaultExchangeContainerLocked(receiverID),
				ProgressByGroupID:   map[int64]exchangeProgress{},
			},
		},
	}
	world.exchangeSessions = append(world.exchangeSessions, session)
	world.broadcastExchangeStateLocked(&world.exchangeSessions[len(world.exchangeSessions)-1])
}

func (world *World) exchangeParticipantByAccountLocked(accountID int64) (*exchangeSession, *exchangeParticipant, *exchangeParticipant, error) {
	_, session, err := world.exchangeSessionByAccountLocked(accountID)
	if err != nil {
		return nil, nil, nil, err
	}
	objectID := world.accountObjectIDs[accountID]
	for index := range session.Participants {
		if session.Participants[index].ObjectID == objectID {
			peerIndex := 1 - index
			return session, &session.Participants[index], &session.Participants[peerIndex], nil
		}
	}
	return nil, nil, nil, errors.New("exchange participant not found")
}

func (world *World) exchangeSessionByAccountLocked(accountID int64) (int, *exchangeSession, error) {
	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return -1, nil, errors.New("controlled object not found")
	}
	for index := range world.exchangeSessions {
		session := &world.exchangeSessions[index]
		if session.Participants[0].ObjectID == objectID || session.Participants[1].ObjectID == objectID {
			return index, session, nil
		}
	}
	return -1, nil, errors.New("exchange session not found")
}

func (world *World) exchangeOwnedContainerLocked(objectID int64, containerID int64) (*data.EquipmentGroup, error) {
	return world.controlledContainerEquipmentLocked(objectID, containerID)
}

// Возвращает первый доступный контейнер участника обмена.
func (world *World) defaultExchangeContainerLocked(objectID int64) int64 {
	containers := world.exchangeOwnedContainersOnObjectLocked(objectID, objectID, map[int64]bool{})
	if len(containers) == 0 {
		return 0
	}
	return containers[0].ID
}

func (world *World) exchangeTaskLocked(participant *exchangeParticipant) (*data.Task, error) {
	if participant.TaskID > 0 {
		if task, ok := world.data.Tasks.Get(participant.TaskID); ok {
			return task, nil
		}
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym(exchangeTaskAcronym)
	if !ok {
		return nil, errors.New("exchange task type not found")
	}
	task, err := world.data.Tasks.Add(&data.Task{
		ControllerEquipmentGroupID: participant.SourceContainerID,
		TaskTypeID:                 taskType.ID,
		BatchCount:                 1,
	})
	if err != nil {
		return nil, err
	}
	participant.TaskID = task.ID
	return task, nil
}

func (world *World) exchangeTaskGroupsLocked(participant *exchangeParticipant) []*data.TaskItemGroup {
	if participant == nil || participant.TaskID <= 0 {
		return nil
	}
	return world.data.TaskItemGroups.GetByTaskID(participant.TaskID)
}

func (world *World) prepareExchangeProgressLocked(session *exchangeSession) {
	for participantIndex := range session.Participants {
		participant := &session.Participants[participantIndex]
		for _, group := range world.exchangeTaskGroupsLocked(participant) {
			if group.IsReadyToExchange {
				continue
			}
			energy := math.Max(1, world.exchangeGroupEnergyLocked(session, group))
			participant.ProgressByGroupID[group.ID] = exchangeProgress{RemainingEnergy: energy, TotalEnergy: energy}
		}
	}
}

func (world *World) stepExchangeMovementLocked(session *exchangeSession, dtSeconds float64) {
	workPower := math.Max(1, world.exchangeWorkPowerLocked(session))
	for participantIndex := range session.Participants {
		participant := &session.Participants[participantIndex]
		for _, group := range world.exchangeTaskGroupsLocked(participant) {
			if group.IsReadyToExchange {
				continue
			}
			progress := participant.ProgressByGroupID[group.ID]
			if progress.TotalEnergy <= 0 {
				progress.TotalEnergy = math.Max(1, world.exchangeGroupEnergyLocked(session, group))
				progress.RemainingEnergy = progress.TotalEnergy
			}
			progress.RemainingEnergy = math.Max(0, progress.RemainingEnergy-workPower*dtSeconds)
			participant.ProgressByGroupID[group.ID] = progress
			if progress.RemainingEnergy <= physics.Epsilon {
				group.IsReadyToExchange = true
			}
			return
		}
	}
}

func (world *World) exchangeSessionReadyLocked(session *exchangeSession) bool {
	for participantIndex := range session.Participants {
		for _, group := range world.exchangeTaskGroupsLocked(&session.Participants[participantIndex]) {
			if !group.IsReadyToExchange {
				return false
			}
		}
	}
	return true
}

func (world *World) finishExchangeSessionLocked(session *exchangeSession) {
	left := &session.Participants[0]
	right := &session.Participants[1]
	world.moveExchangeGroupsToContainerLocked(left, right.ReceiverContainerID)
	world.moveExchangeGroupsToContainerLocked(right, left.ReceiverContainerID)
	world.deleteExchangeTaskLocked(left)
	world.deleteExchangeTaskLocked(right)
}

func (world *World) moveExchangeGroupsToContainerLocked(participant *exchangeParticipant, targetContainerID int64) {
	for _, group := range world.exchangeTaskGroupsLocked(participant) {
		_ = world.addItemModelToContainerLocked(targetContainerID, group.ItemModelID, group.Count)
	}
}

func (world *World) cancelExchangeSessionLocked(session *exchangeSession) {
	for participantIndex := range session.Participants {
		participant := &session.Participants[participantIndex]
		groups := world.exchangeTaskGroupsLocked(participant)
		for _, container := range world.exchangeReturnContainersLocked(participant) {
			if world.containerCanAcceptTaskReserveLocked(container.ID, groups) {
				world.moveExchangeGroupsToContainerLocked(participant, container.ID)
				break
			}
		}
		world.deleteExchangeTaskLocked(participant)
	}
}

func (world *World) exchangeReturnContainersLocked(participant *exchangeParticipant) []*data.EquipmentGroup {
	result := make([]*data.EquipmentGroup, 0)
	seen := make(map[int64]bool)
	sourceObjectID := participant.ObjectID
	if source, ok := world.data.EquipmentGroups.Get(participant.SourceContainerID); ok {
		sourceObjectID = source.CosmicObjectID
		result = append(result, source)
		seen[source.ID] = true
	}
	for _, container := range world.exchangeOwnedContainersOnObjectLocked(participant.ObjectID, sourceObjectID, seen) {
		result = append(result, container)
		seen[container.ID] = true
	}
	participantObject, ok := world.data.CosmicObjects.Get(participant.ObjectID)
	if !ok {
		return result
	}
	mainID := world.clusterMainObjectIDLocked(participantObject)
	for _, objectID := range world.clusterObjectIDsLocked(mainID) {
		if objectID == sourceObjectID {
			continue
		}
		for _, container := range world.exchangeOwnedContainersOnObjectLocked(participant.ObjectID, objectID, seen) {
			result = append(result, container)
			seen[container.ID] = true
		}
	}
	return result
}

func (world *World) exchangeOwnedContainersOnObjectLocked(controlledObjectID int64, objectID int64, seen map[int64]bool) []*data.EquipmentGroup {
	containers := make([]*data.EquipmentGroup, 0)
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group == nil || seen[group.ID] || !world.equipmentGroupIsContainerLocked(group) {
			continue
		}
		if err := world.ensureControlledClusterEquipmentLocked(controlledObjectID, group.CosmicObjectID); err != nil {
			continue
		}
		containers = append(containers, group)
	}
	sort.Slice(containers, func(left int, right int) bool { return containers[left].ID < containers[right].ID })
	return containers
}

func (world *World) deleteExchangeTaskLocked(participant *exchangeParticipant) {
	if participant.TaskID <= 0 {
		return
	}
	world.data.TaskItemGroups.DeleteByTaskID(participant.TaskID)
	world.data.Tasks.Delete(participant.TaskID)
	participant.TaskID = 0
}

func (world *World) exchangeGroupEnergyLocked(session *exchangeSession, group *data.TaskItemGroup) float64 {
	source, _ := world.data.CosmicObjects.Get(session.SenderCosmicObjectID)
	target, _ := world.data.CosmicObjects.Get(session.ReceiverCosmicObjectID)
	if source == nil || target == nil {
		return group.Count
	}
	itemModel, ok := world.data.ItemModels.Get(group.ItemModelID)
	if !ok {
		return group.Count
	}
	distance := math.Hypot(source.X-target.X, source.Y-target.Y)
	return math.Max(1, itemModel.Mass*group.Count*math.Max(1, distance))
}

func (world *World) exchangeWorkPowerLocked(session *exchangeSession) float64 {
	var power float64
	objectIDs := make(map[int64]bool)
	for participantIndex := range session.Participants {
		object, ok := world.data.CosmicObjects.Get(session.Participants[participantIndex].ObjectID)
		if !ok {
			continue
		}
		for _, objectID := range world.clusterObjectIDsLocked(world.clusterMainObjectIDLocked(object)) {
			objectIDs[objectID] = true
		}
	}
	for objectID := range objectIDs {
		for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
			model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
			if ok && group.Enabled && group.EnabledCount > 0 {
				power += model.ConsumingPower * float64(group.EnabledCount)
			}
		}
	}
	return power
}

func (world *World) exchangeObjectIsBusyLocked(objectID int64) bool {
	for _, request := range world.exchangeRequests {
		if request.SenderCosmicObjectID == objectID || request.ReceiverCosmicObjectID == objectID {
			return true
		}
	}
	for _, session := range world.exchangeSessions {
		if session.Participants[0].ObjectID == objectID || session.Participants[1].ObjectID == objectID {
			return true
		}
	}
	return false
}

func (world *World) exchangeClusterIsBusyLocked(objectID int64) bool {
	object, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return false
	}
	if world.exchangeObjectIsBusyLocked(objectID) {
		return true
	}
	mainID := object.ClusterMainCosmicObjectID
	if mainID <= 0 {
		return false
	}
	for _, clusterObjectID := range world.clusterObjectIDsLocked(mainID) {
		if world.exchangeObjectIsBusyLocked(clusterObjectID) {
			return true
		}
	}
	return false
}

func (world *World) exchangePausesControllerLocked(controllerID int64) bool {
	controller, ok := world.data.EquipmentGroups.Get(controllerID)
	if !ok {
		return false
	}
	for index := range world.exchangeSessions {
		session := &world.exchangeSessions[index]
		if !session.Moving {
			continue
		}
		for participantIndex := range session.Participants {
			participantObject, ok := world.data.CosmicObjects.Get(session.Participants[participantIndex].ObjectID)
			if !ok {
				continue
			}
			if participantObject.ClusterMainCosmicObjectID > 0 {
				if controllerObject, ok := world.data.CosmicObjects.Get(controller.CosmicObjectID); ok && controllerObject.ClusterMainCosmicObjectID == participantObject.ClusterMainCosmicObjectID {
					return true
				}
				continue
			}
			if controller.CosmicObjectID == participantObject.ID {
				return true
			}
		}
	}
	return false
}

func (world *World) addExchangeWindowEventsLocked(kind string, senderID int64, receiverID int64, duration float64) {
	world.exchangeEvents = append(world.exchangeEvents,
		game.ExchangeEvent{Type: "exchangeEvent", Kind: kind, Role: "sender", Duration: duration, ObjectIDs: []int64{senderID}},
		game.ExchangeEvent{Type: "exchangeEvent", Kind: kind, Role: "receiver", Duration: duration, ObjectIDs: []int64{receiverID}},
	)
}

func (world *World) addExchangeNotificationLocked(objectIDs []int64, message string) {
	world.exchangeEvents = append(world.exchangeEvents, game.ExchangeEvent{
		Type:      "exchangeEvent",
		Kind:      "exchangeNotification",
		Message:   message,
		ObjectIDs: objectIDs,
	})
}

func (world *World) broadcastExchangeStateLocked(session *exchangeSession) {
	world.exchangeEvents = append(world.exchangeEvents,
		game.ExchangeEvent{Type: "exchangeEvent", Kind: "exchangeState", Role: "sender", State: world.exchangeStateForLocked(session, 0), ObjectIDs: []int64{session.Participants[0].ObjectID}},
		game.ExchangeEvent{Type: "exchangeEvent", Kind: "exchangeState", Role: "receiver", State: world.exchangeStateForLocked(session, 1), ObjectIDs: []int64{session.Participants[1].ObjectID}},
	)
}

func (world *World) exchangeStateForLocked(session *exchangeSession, participantIndex int) game.ExchangeState {
	participant := &session.Participants[participantIndex]
	peer := &session.Participants[1-participantIndex]
	return game.ExchangeState{
		SelfObjectID:            participant.ObjectID,
		OtherObjectID:           peer.ObjectID,
		SelfNickname:            world.nicknameForObjectLocked(participant.ObjectID),
		OtherNickname:           world.nicknameForObjectLocked(peer.ObjectID),
		SelfReceiverContainerID: participant.ReceiverContainerID,
		SelfSourceContainerID:   participant.SourceContainerID,
		SelfConfirmed:           participant.Confirmed,
		OtherConfirmed:          peer.Confirmed,
		NotEnoughSpace:          participant.NotEnoughSpace,
		SelfQueue:               world.exchangeQueueItemsLocked(participant),
		OtherQueue:              world.exchangeQueueItemsLocked(peer),
	}
}

func (world *World) exchangeQueueItemsLocked(participant *exchangeParticipant) []game.ExchangeQueueItem {
	items := make([]game.ExchangeQueueItem, 0)
	for _, group := range world.exchangeTaskGroupsLocked(participant) {
		progress := participant.ProgressByGroupID[group.ID]
		value := 0.0
		if progress.TotalEnergy > physics.Epsilon {
			value = 1 - progress.RemainingEnergy/progress.TotalEnergy
		}
		items = append(items, game.ExchangeQueueItem{TaskItemGroupID: group.ID, ItemModelID: group.ItemModelID, Count: group.Count, Progress: value, IsReady: group.IsReadyToExchange})
	}
	return items
}

func (world *World) nicknameForObjectLocked(objectID int64) string {
	object, ok := world.data.CosmicObjects.Get(objectID)
	if !ok || object.OwnerCharacterID <= 0 {
		return ""
	}
	character, ok := world.data.Characters.Get(object.OwnerCharacterID)
	if !ok {
		return ""
	}
	account, ok := world.data.Accounts.Get(character.AccountID)
	if !ok {
		return ""
	}
	return account.Nickname
}
