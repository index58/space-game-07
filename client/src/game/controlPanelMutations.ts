import type { ControlPanelMutationAck, ControlPanelMutationRef, CosmicObject, EquipmentGroup } from "../network/protocol";

export type PendingValue<T> = ControlPanelMutationRef & {
  // Значение, которое игрок уже отправил на сервер, но еще не увидел подтвержденным в снимке.
  value: T;
};

export type ControlPanelPendingState = {
  // Ожидающие изменения свойств управляемого объекта.
  object: {
    // Ожидающее изменение признака включения объекта.
    enabled?: PendingValue<boolean>;
    // Ожидающее изменение пользовательского названия объекта.
    title?: PendingValue<string>;
  };
  // Ожидающие изменения групп оборудования по ID группы.
  equipment: Record<number, {
    // Ожидающее изменение признака включения группы.
    enabled?: PendingValue<boolean>;
    // Ожидающее изменение количества включенных единиц.
    enabledCount?: PendingValue<number>;
  }>;
};

export const emptyControlPanelPendingState = (): ControlPanelPendingState => ({
  object: {},
  equipment: {},
});

// Накладывает ожидающие изменения панели на объект из последнего серверного снимка.
export const applyControlPanelPendingToObject = (object: CosmicObject | null, pending: ControlPanelPendingState): CosmicObject | null => {
  if (!object) {
    return null;
  }

  return {
    ...object,
    Enabled: pending.object.enabled?.value ?? object.Enabled,
    Title: pending.object.title?.value ?? object.Title,
  };
};

// Накладывает ожидающие изменения панели на группы оборудования из последнего серверного снимка.
export const applyControlPanelPendingToEquipmentGroups = (groups: EquipmentGroup[], pending: ControlPanelPendingState): EquipmentGroup[] => {
  return groups.map((group) => {
    const groupPending = pending.equipment[group.ID];
    if (!groupPending) {
      return group;
    }

    return {
      ...group,
      Enabled: groupPending.enabled?.value ?? group.Enabled,
      EnabledCount: groupPending.enabledCount?.value ?? group.EnabledCount,
    };
  });
};

// Удаляет ожидающие изменения, которые сервер уже подтвердил водяным знаком мутаций.
export const pruneControlPanelPending = (pending: ControlPanelPendingState, ack: ControlPanelMutationAck | undefined): ControlPanelPendingState => {
  if (!ack) {
    return pending;
  }

  return {
    object: {
      enabled: keepPendingValue(pending.object.enabled, ack),
      title: keepPendingValue(pending.object.title, ack),
    },
    equipment: Object.fromEntries(
      Object.entries(pending.equipment)
        .map(([groupId, groupPending]) => [groupId, {
          enabled: keepPendingValue(groupPending.enabled, ack),
          enabledCount: keepPendingValue(groupPending.enabledCount, ack),
        }] as const)
        .filter(([, groupPending]) => groupPending.enabled || groupPending.enabledCount),
    ),
  };
};

// Удаляет ожидающее изменение, которое сервер явно отклонил.
export const rejectControlPanelPending = (pending: ControlPanelPendingState, rejected: { clientSessionId: string; mutationSeq: number }): ControlPanelPendingState => {
  return {
    object: {
      enabled: rejectPendingValue(pending.object.enabled, rejected),
      title: rejectPendingValue(pending.object.title, rejected),
    },
    equipment: Object.fromEntries(
      Object.entries(pending.equipment)
        .map(([groupId, groupPending]) => [groupId, {
          enabled: rejectPendingValue(groupPending.enabled, rejected),
          enabledCount: rejectPendingValue(groupPending.enabledCount, rejected),
        }] as const)
        .filter(([, groupPending]) => groupPending.enabled || groupPending.enabledCount),
    ),
  };
};

// Оставляет значение только если серверный ack еще не дошел до его номера.
const keepPendingValue = <T>(value: PendingValue<T> | undefined, ack: ControlPanelMutationAck): PendingValue<T> | undefined => {
  if (!value || value.sessionId !== ack.sessionId) {
    return value;
  }

  return value.seq <= ack.lastAppliedSeq ? undefined : value;
};

// Оставляет значение, если отказ сервера относится к другой мутации.
const rejectPendingValue = <T>(value: PendingValue<T> | undefined, rejected: { clientSessionId: string; mutationSeq: number }): PendingValue<T> | undefined => {
  if (!value || value.sessionId !== rejected.clientSessionId) {
    return value;
  }

  return value.seq === rejected.mutationSeq ? undefined : value;
};
