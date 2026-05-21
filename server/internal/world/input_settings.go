package world

import (
	"errors"

	"space-game-07-server/internal/data"
)

// AccountInputSettings РІРѕР·РІСЂР°С‰Р°РµС‚ РєРѕРїРёРё РїСЂРёРІСЏР·РѕРє РІРІРѕРґР° С‚РѕР»СЊРєРѕ РґР»СЏ СѓРєР°Р·Р°РЅРЅРѕРіРѕ Р°РєРєР°СѓРЅС‚Р°.
func (world *World) AccountInputSettings(accountID int64) []data.AccountActionInputSetting {
	world.mu.Lock()
	defer world.mu.Unlock()

	if world.data.AccountActionInputSettings == nil {
		return nil
	}
	settings := world.data.AccountActionInputSettings.GetByAccount(accountID)
	result := make([]data.AccountActionInputSetting, 0, len(settings))
	for _, setting := range settings {
		result = append(result, *setting)
	}
	return result
}

// SaveAccountInputSettings РїСЂРѕРІРµСЂСЏРµС‚ Рё Р·Р°РјРµРЅСЏРµС‚ РїСЂРёРІСЏР·РєРё РІРІРѕРґР° С‚РµРєСѓС‰РµРіРѕ Р°РєРєР°СѓРЅС‚Р° РІ РїР°РјСЏС‚Рё РјРёСЂР°.
func (world *World) SaveAccountInputSettings(accountID int64, settings []data.AccountActionInputSetting) ([]data.AccountActionInputSetting, error) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if _, ok := world.data.Accounts.Get(accountID); !ok {
		return nil, errors.New("account not found")
	}
	if world.data.ActionTypes == nil || world.data.InputEventTypes == nil {
		return nil, errors.New("input reference data is not loaded")
	}
	if world.data.AccountActionInputSettings == nil {
		world.data.AccountActionInputSettings = data.NewAccountActionInputSettings()
	}

	actionIDs := map[int64]struct{}{}
	eventIDs := map[int64]struct{}{}
	clean := make([]data.AccountActionInputSetting, 0, len(settings))
	for _, setting := range settings {
		if _, ok := world.data.ActionTypes.Get(setting.ActionTypeID); !ok {
			return nil, errors.New("input action type not found")
		}
		if _, ok := world.data.InputEventTypes.Get(setting.InputEventTypeID); !ok {
			return nil, errors.New("input event type not found")
		}
		if _, exists := actionIDs[setting.ActionTypeID]; exists {
			return nil, errors.New("input action type is duplicated")
		}
		if _, exists := eventIDs[setting.InputEventTypeID]; exists {
			return nil, errors.New("input event type is duplicated")
		}
		actionIDs[setting.ActionTypeID] = struct{}{}
		eventIDs[setting.InputEventTypeID] = struct{}{}
		clean = append(clean, data.AccountActionInputSetting{
			AccountID:        accountID,
			ActionTypeID:     setting.ActionTypeID,
			InputEventTypeID: setting.InputEventTypeID,
		})
	}

	if err := world.data.AccountActionInputSettings.ReplaceForAccount(accountID, clean); err != nil {
		return nil, err
	}
	return world.accountInputSettingsLocked(accountID), nil
}

// accountInputSettingsLocked РІРѕР·РІСЂР°С‰Р°РµС‚ РєРѕРїРёРё РїСЂРёРІСЏР·РѕРє РІРІРѕРґР° РїРѕРґ СѓР¶Рµ РІР·СЏС‚С‹Рј mutex РјРёСЂР°.
func (world *World) accountInputSettingsLocked(accountID int64) []data.AccountActionInputSetting {
	if world.data.AccountActionInputSettings == nil {
		return nil
	}
	settings := world.data.AccountActionInputSettings.GetByAccount(accountID)
	result := make([]data.AccountActionInputSetting, 0, len(settings))
	for _, setting := range settings {
		result = append(result, *setting)
	}
	return result
}
