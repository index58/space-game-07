import { createMemo, For, Match, Show, Switch, type Accessor, type JSX } from "solid-js";
import { Portal } from "solid-js/web";
import { formatNumber } from "../domain/format";
import type { CosmicObjectModelReference, EquipmentGroup, ItemModelReference } from "../network/protocol";
import type { ControlPanelEquipmentSubTabValue, ControlPanelTabValue, GameUiState, SettingsTabValue } from "./gameUiState";
import { getDebugOverlayLines } from "./debugOverlay";
import { getObjectIndicators, type ObjectIndicatorView } from "./objectIndicators";
import { getMinimapView, type MinimapPointView } from "./minimap";
import { getPilotToolbarView, type PilotToolSlotView } from "./pilotToolbar";
import { getInformationPanelView, type InformationPanelRow } from "./informationPanel";
import { getInputEventOptions, getInputSettingsLeftColumnRowCount, getInputSettingsRows, type InputSettingsRow } from "./inputSettings";
import { Button, Checkbox, ContextMenu, Dropdown, EditControl, ListBox, Modal, NumericStepper, RadioGroup, Scrollbar, Slider, Splitter, Tabs, TextInput, Tooltip, TreeView, VirtualList } from "../ui-kit/components";

type HudPanelPosition = "left-bottom" | "left-middle" | "bottom-center" | "right-middle" | "right-bottom" | "left-top";

type GameUiProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type HudPanelProps = {
  // Расположение панели относительно игрового экрана.
  position: HudPanelPosition;
  // Частный CSS-класс содержимого конкретной панели.
  className: string;
  // Доступное название панели для браузерного дерева.
  ariaLabel?: string;
  // Стабильный идентификатор панели, если он нужен внешнему коду или тестам.
  id?: string;
  // Внутреннее содержимое панели.
  children: JSX.Element;
};

type ObjectIndicatorProps = {
  // Данные одной строки панели основных показателей.
  indicator: ObjectIndicatorView;
};

// Корневой компонент всех текущих UI-слоёв поверх Phaser canvas.
export const GameUi = (props: GameUiProps) => (
  <>
    <ObjectIndicatorsPanel selfObject={props.state().selfObject} />
    <ChatPanel state={props.state} />
    <InformationPanel state={props.state} />
    <PilotToolbarPanel state={props.state} />
    <MinimapPanel state={props.state} />
    <DebugOverlay state={props.state} />
    <UiKitShowcase state={props.state} />
    <SettingsModal state={props.state} />
    <ControlPanelModal state={props.state} />
    <GameCursor state={props.state} />
  </>
);

// Задаёт единый технический каркас для всех игровых HUD-панелей.
const HudPanel = (props: HudPanelProps) => (
  <section
    id={props.id}
    class={`hud-panel hud-panel--${props.position} ${props.className}`}
    aria-label={props.ariaLabel}
  >
    {props.children}
  </section>
);

type ObjectIndicatorsPanelProps = {
  // Посещаемый объект игрока, если он уже получен.
  selfObject: GameUiState["selfObject"];
};

// Показывает основные показатели посещаемого объекта в левой нижней части экрана.
const ObjectIndicatorsPanel = (props: ObjectIndicatorsPanelProps) => (
  <Show when={props.selfObject}>
    {(selfObject) => (
      <HudPanel position="left-bottom" className="object-indicators" ariaLabel="Основные показатели посещаемого объекта">
        <For each={getObjectIndicators(selfObject())}>
          {(indicator) => <ObjectIndicator indicator={indicator} />}
        </For>
      </HudPanel>
    )}
  </Show>
);

// Рисует одну строку со значком, полосой и числовым значением.
const ObjectIndicator = (props: ObjectIndicatorProps) => (
  <div class="object-indicator" title={props.indicator.title}>
    <div class={`object-indicator__icon object-indicator__icon--${props.indicator.acronym}`}>
      <IndicatorIcon acronym={props.indicator.acronym} />
    </div>
    <div class="object-indicator__bar">
      <div class="object-indicator__fill" style={{ width: `${props.indicator.fillPercent}%` }} />
      <div class="object-indicator__value">{props.indicator.valueText}</div>
    </div>
  </div>
);

type IndicatorIconProps = {
  // Стабильный строковый идентификатор значка.
  acronym: ObjectIndicatorView["acronym"];
};

// Выбирает векторный значок для конкретного показателя.
const IndicatorIcon = (props: IndicatorIconProps) => (
  <svg viewBox="0 0 24 24" aria-hidden="true">
    <Switch>
      <Match when={props.acronym === "Armor"}>
        <path d="M12 3l7 3v5c0 5-3 8-7 10-4-2-7-5-7-10V6l7-3z" />
      </Match>
      <Match when={props.acronym === "Power"}>
        <path d="M13 2L5 13h6l-1 9 9-13h-6l1-7z" />
      </Match>
      <Match when={props.acronym === "Fuel"}>
        <path d="M7 3h8v18H7z" />
        <path d="M9 6h4v4H9z" />
        <path d="M10 8h2" />
        <path d="M15 8h2l2 3v7c0 1-1 2-2 2h-2" />
        <path d="M19 11h-2" />
      </Match>
    </Switch>
  </svg>
);

type MinimapPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type InformationPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type InformationPanelRowProps = {
  // Строка с подписью и значением для информационной панели.
  row: InformationPanelRow;
};

type PilotToolbarPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type ChatPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

// Значок нужен только каналам, где короткий тип помогает отличить общий чат от личного диалога.
const chatTabMarkerAcronyms = new Set(["Server", "Clan", "Alliance", "Solo"]);

// Возвращает короткую подпись значка только для типов чата, у которых он должен быть видимым.
const getChatTabMarker = (communityTypeAcronym: string): string | undefined =>
  chatTabMarkerAcronyms.has(communityTypeAcronym) ? communityTypeAcronym.slice(0, 1) : undefined;

type GameCursorProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type UiKitShowcaseProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type SettingsModalProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type SettingsInputRowsProps = {
  // Строки, которые должны быть показаны в одной половине вкладки ввода.
  rows: InputSettingsRow[];
  // Варианты событий ввода для выпадающих списков.
  options: ReturnType<typeof getInputEventOptions>;
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type ControlPanelModalProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type GameWindowLayerProps = {
  // Вид окна, который задаёт необходимые отличия поверх общего каркаса.
  variant: "settings" | "showcase" | "control-panel";
  // Содержимое окна в общем экранном слое.
  children: JSX.Element;
};

type ControlPanelObjectRow = {
  // Видимая подпись характеристики или поля.
  label: string;
  // Текстовое значение, показанное справа от подписи.
  value: string;
};

type ControlPanelEquipmentView = {
  // Группа оборудования из серверного снимка.
  group: EquipmentGroup;
  // Видимое название модели оборудования.
  modelTitle: string;
  // Справочная модель установленного предмета.
  itemModel: ItemModelReference | undefined;
};

type ControlPanelEquipmentInfoRow = {
  // Видимая подпись характеристики выбранного оборудования.
  label: string;
  // Текстовое значение характеристики выбранного оборудования.
  value: string;
};

type GameFormRowLabelProps = {
  // Дополнительный класс конкретной строки формы.
  className?: string;
  // Текст или содержимое общей подписи строки.
  children: JSX.Element;
};

const settingsTabs: Array<{ value: SettingsTabValue; label: string }> = [
  { value: "video", label: "Видео" },
  { value: "audio", label: "Аудио" },
  { value: "input", label: "Ввод" },
];

const controlPanelTabs: Array<{ value: ControlPanelTabValue; label: string }> = [
  { value: "object", label: "Объект" },
  { value: "equipment", label: "Оборудование" },
  { value: "pilotTools", label: "Инструменты пилота" },
  { value: "schemas", label: "Схемы" },
  { value: "blueprints", label: "Чертежи" },
  { value: "map", label: "Карта" },
];

const controlPanelEquipmentTabs: Array<{ value: ControlPanelEquipmentSubTabValue; label: string }> = [
  { value: "setup", label: "Настройка" },
  { value: "usage", label: "Использование" },
];

// Задаёт общий экранный шаблон для всех игровых модальных окон.
const GameWindowLayer = (props: GameWindowLayerProps) => (
  <div class={`game-window-layer game-window-layer--${props.variant}`}>
    {props.children}
  </div>
);

// Рисует левую часть строк игровых форм по единому шаблону окна настроек.
const GameFormRowLabel = (props: GameFormRowLabelProps) => (
  <div class={`game-form-row-label ${props.className ?? ""}`}>{props.children}</div>
);

type PilotToolSlotProps = {
  // Данные одной ячейки панели инструментов пилота.
  slot: PilotToolSlotView;
};

type PilotToolbarReadyState = {
  // Посещаемый объект, для которого строится панель пилота.
  selfObject: NonNullable<GameUiState["selfObject"]>;
  // Справочники клиента для определения инструментов пилота.
  referenceData: NonNullable<GameUiState["referenceData"]>;
};

// Возвращает данные панели пилота только после получения объекта и справочников.
const getPilotToolbarReadyState = (state: GameUiState): PilotToolbarReadyState | null => {
  if (!state.selfObject || !state.referenceData) {
    return null;
  }
  return {
    selfObject: state.selfObject,
    referenceData: state.referenceData,
  };
};

// Показывает краткую информацию об объекте, на который смотрит нос корабля.
const InformationPanel = (props: InformationPanelProps) => {
  const view = () => {
    const state = props.state();
    if (!state.selfObject || !state.referenceData) {
      return null;
    }
    return getInformationPanelView({
      selfObject: state.selfObject,
      objects: state.objects,
      referenceData: state.referenceData,
    });
  };

  return (
    <Show when={view()}>
      {(panel) => (
        <HudPanel position="right-middle" className="information-panel" ariaLabel="Информационная панель">
          <For each={panel().rows}>
            {(row) => <InformationPanelRow row={row} />}
          </For>
        </HudPanel>
      )}
    </Show>
  );
};

// Рисует одну строку информационной панели.
const InformationPanelRow = (props: InformationPanelRowProps) => (
  <div class="information-panel__row">
    <div class="information-panel__label">{props.row.label}</div>
    <div class="information-panel__value">{props.row.value}</div>
  </div>
);

// Показывает доступные вкладки, последние строки истории и локальную строку ввода.
const ChatPanel = (props: ChatPanelProps) => {
  const selectedTab = () => props.state().chatState?.tabs.find((tab) => tab.chatId === props.state().chatState?.selectedChatId) ?? null;
  const chatErrorAnimationName = () => props.state().chatErrorSeq % 2 === 0 ? "chat-error-fade-even" : "chat-error-fade-odd";
  const chatSelectionStart = () => Math.min(props.state().chatSelectionStart, props.state().chatSelectionEnd);
  const chatSelectionEnd = () => Math.max(props.state().chatSelectionStart, props.state().chatSelectionEnd);
  const chatTabs = () => (props.state().chatState?.tabs ?? []).map((chatTab) => ({
    value: String(chatTab.chatId),
    label: chatTab.title,
    marker: getChatTabMarker(chatTab.communityTypeAcronym),
    badge: (chatTab.unreadCount ?? 0) > 0 ? chatTab.unreadCount : undefined,
  }));

  return (
    <Show when={props.state().chatState && selectedTab()}>
      {(tab) => (
        <HudPanel position="left-middle" className="chat-panel" ariaLabel="Чат">
          <div class="chat-messages">
            <div
              class="chat-messages__content"
              style={{ transform: `translateY(${props.state().chatScroll.contentOffsetPx}px)` }}
            >
              <For each={tab().messages}>
                {(message) => (
                  <div class="chat-message" style={{ color: `#${message.color || "d8f3ff"}` }}>
                    <span class="chat-message__sender">{message.senderNickname}</span>
                    <span class="chat-message__separator">: </span>
                    <span class="chat-message__text">{message.text}</span>
                  </div>
                )}
              </For>
            </div>
            <Show when={props.state().chatScroll.visible}>
              <Scrollbar
                id="chat-history-scrollbar"
                className="chat-scrollbar"
                thumbTopPercent={props.state().chatScroll.thumbTopPercent}
                thumbHeightPercent={props.state().chatScroll.thumbHeightPercent}
                dragging={props.state().chatScroll.dragging}
              />
            </Show>
          </div>
          <Show when={props.state().chatError}>
            {(error) => <div class="chat-error" style={{ "animation-name": chatErrorAnimationName() }}>{error()}</div>}
          </Show>
          <TextInput
            id="chat-input"
            className="chat-input"
            text={props.state().chatInputText}
            selectionStart={chatSelectionStart()}
            selectionEnd={chatSelectionEnd()}
            focused={props.state().chatInputFocused}
          />
          <Tabs
            id="chat-tabs"
            itemIdPrefix="chat-tab"
            className="chat-tabs"
            itemClassName="chat-tab"
            selectedValue={String(tab().chatId)}
            tabs={chatTabs()}
          />
          <Show when={props.state().chatContextMenu}>
            {(menu) => (
              <div
                id="chat-context-menu"
                data-ui-kind="menu"
                class="ui-kit-control chat-context-menu"
                style={{ left: `${menu().x}px`, top: `${menu().y}px` }}
              >
                <div class={`chat-context-menu__item ${menu().communityTypeAcronym === "Duo" ? "" : "is-disabled"}`}>Закрыть</div>
              </div>
            )}
          </Show>
        </HudPanel>
      )}
    </Show>
  );
};

// Рисует серый игровой указатель поверх HUD, когда системный указатель захвачен игрой.
const GameCursor = (props: GameCursorProps) => (
  <Show when={props.state().gameCursor.visible}>
    <Portal>
      <div class="game-cursor" style={{ left: `${props.state().gameCursor.x}px`, top: `${props.state().gameCursor.y}px` }} />
    </Portal>
  </Show>
);

// Показывает отладочную витрину всех базовых контролов UI Kit.
const UiKitShowcase = (props: UiKitShowcaseProps) => (
  <Show when={props.state().uiKitShowcaseVisible}>
    <GameWindowLayer variant="showcase">
      <Modal id="ui-kit-showcase-modal" title="UI Kit">
        <div class="ui-kit-showcase">
          <div class="ui-kit-showcase__grid">
            <Button id="ui-kit-demo-button" label={`Button ${props.state().uiKitDemoState.buttonClicks}`} state="hovered" />
            <Button id="ui-kit-demo-icon-button" label="*" state="focused" />
            <Checkbox id="ui-kit-demo-checkbox" label="Checkbox" checked={props.state().uiKitDemoState.checkboxChecked} />
            <RadioGroup id="ui-kit-demo-radio" value={props.state().uiKitDemoState.radioValue} options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
            <Dropdown id="ui-kit-demo-select" label="" selectedValue={props.state().uiKitDemoState.dropdownValue} open={props.state().uiKitDemoState.dropdownOpen} options={[{ value: "one", label: "One" }, { value: "two", label: "Two" }]} />
            <ListBox id="ui-kit-demo-list" selectedValue={props.state().uiKitDemoState.listValue} items={[{ value: "1", label: "One" }, { value: "2", label: "Two" }]} />
            <TreeView id="ui-kit-demo-tree" selectedValue={props.state().uiKitDemoState.treeValue} nodes={[{ value: "root", label: "Root", children: [{ value: "child", label: "Child" }] }]} />
            <VirtualList id="ui-kit-demo-virtual-list" startIndex={props.state().uiKitDemoState.virtualStartIndex} items={[{ value: "20", label: "Item 20" }, { value: "21", label: "Item 21" }]} />
            <Tabs id="ui-kit-demo-tabs" selectedValue={props.state().uiKitDemoState.tabValue} tabs={[{ value: "one", label: "One" }, { value: "two", label: "Two", badge: 3 }]} />
            <EditControl id="ui-kit-demo-edit" text={props.state().uiKitDemoState.editText} selectionStart={props.state().uiKitDemoState.editSelectionStart} selectionEnd={props.state().uiKitDemoState.editSelectionEnd} focused={true} />
            <Scrollbar id="ui-kit-demo-scrollbar" thumbTopPercent={props.state().uiKitDemoState.scrollbarTopPercent} thumbHeightPercent={45} dragging={props.state().uiKitDemoState.scrollbarDrag !== null} />
            <Slider id="ui-kit-demo-slider" value={props.state().uiKitDemoState.sliderValue} min={0} max={100} />
            <NumericStepper id="ui-kit-demo-stepper" value={props.state().uiKitDemoState.stepperValue} />
            <Splitter id="ui-kit-demo-splitter" vertical={props.state().uiKitDemoState.splitterVertical} />
            <Show when={props.state().uiKitDemoState.menuOpen}>
              <ContextMenu id="ui-kit-demo-menu" items={[{ value: "close", label: "Close" }]} />
            </Show>
            <Show when={props.state().uiKitDemoState.tooltipVisible}>
              <Tooltip id="ui-kit-demo-tooltip" text="Tooltip" />
            </Show>
          </div>
        </div>
      </Modal>
    </GameWindowLayer>
  </Show>
);

// Показывает окно настроек аккаунта поверх игрового HUD.
const SettingsModal = (props: SettingsModalProps) => {
  const rows = createMemo(() => getInputSettingsRows(props.state().referenceData, props.state().inputSettingsValues));
  const leftRowCount = createMemo(() => getInputSettingsLeftColumnRowCount(rows().length));
  const leftRows = createMemo(() => rows().slice(0, leftRowCount()));
  const rightRows = createMemo(() => rows().slice(leftRowCount()));
  const options = createMemo(() => getInputEventOptions(props.state().referenceData));
  return (
    <Show when={props.state().settingsVisible}>
      <GameWindowLayer variant="settings">
        <Modal id="settings-modal" title="Настройки">
          <div class="settings-modal">
            <Tabs id="settings-tabs" itemIdPrefix="settings-tab" align="center" className="settings-tabs" selectedValue={props.state().selectedSettingsTab} tabs={settingsTabs} />
            <Switch>
              <Match when={props.state().selectedSettingsTab === "video"}>
                <div class="settings-empty-page" />
              </Match>
              <Match when={props.state().selectedSettingsTab === "audio"}>
                <div class="settings-empty-page" />
              </Match>
              <Match when={props.state().selectedSettingsTab === "input"}>
                <div class="settings-input-table">
                  <div class="settings-input-table__left">
                    <SettingsInputRows rows={leftRows()} options={options()} state={props.state} />
                  </div>
                  <div class="settings-input-table__right">
                    <SettingsInputRows rows={rightRows()} options={options()} state={props.state} />
                  </div>
                  <Show when={props.state().inputSettingsScroll.visible}>
                    <Scrollbar
                      id="settings-input-scrollbar"
                      className="settings-input-scrollbar"
                      thumbTopPercent={props.state().inputSettingsScroll.thumbTopPercent}
                      thumbHeightPercent={props.state().inputSettingsScroll.thumbHeightPercent}
                      dragging={props.state().inputSettingsScroll.dragging}
                    />
                  </Show>
                </div>
              </Match>
            </Switch>
            <div class="settings-modal__footer">
              <div class="settings-modal__actions">
                <Button id="settings-save-button" label={props.state().inputSettingsSaving ? "Сохранение" : "Сохранить"} />
                <Button id="settings-cancel-button" label="Отмена" />
              </div>
              <Show when={props.state().inputSettingsError}>
                {(error) => <div class="settings-modal__error">{error()}</div>}
              </Show>
            </div>
          </div>
        </Modal>
      </GameWindowLayer>
    </Show>
  );
};

// Отрисовывает строки одной половины вкладки ввода с общим вертикальным смещением прокрутки.
const SettingsInputRows = (props: SettingsInputRowsProps) => (
  <div class="settings-input-table__content" style={{ transform: `translateY(-${props.state().inputSettingsScroll.contentOffsetPx}px)` }}>
    <For each={props.rows}>
      {(row) => (
        <div class="settings-input-row">
          <GameFormRowLabel className="settings-input-row__action">{row.actionTitle}</GameFormRowLabel>
          <Dropdown
            id={`settings-input-select-${row.actionTypeId}`}
            selectedValue={String(row.inputEventTypeId)}
            open={props.state().openInputSettingsActionId === row.actionTypeId}
            options={props.options}
            menuScroll={props.state().inputSettingsDropdownScroll}
          />
        </div>
      )}
    </For>
  </div>
);

// Показывает модальную панель управления текущим объектом поверх игрового HUD.
const ControlPanelModal = (props: ControlPanelModalProps) => {
  const modelTitle = createMemo(() => getControlPanelModelTitle(props.state()));
  const rows = createMemo(() => getControlPanelObjectRows(props.state()));
  const equipmentGroups = createMemo(() => getControlPanelEquipmentGroups(props.state()));
  const selectedEquipment = createMemo(() => getSelectedControlPanelEquipment(equipmentGroups(), props.state().selectedControlPanelEquipmentGroupId));
  const selectedEquipmentEnabled = createMemo(() => getControlPanelEquipmentEnabled(props.state(), selectedEquipment()));
  const selectedEquipmentEnabledCount = createMemo(() => getControlPanelEquipmentEnabledCount(props.state(), selectedEquipment()));
  const selectedEquipmentRows = createMemo(() => getControlPanelEquipmentInfoRows(selectedEquipment()));

  return (
    <Show when={props.state().controlPanelVisible}>
      <GameWindowLayer variant="control-panel">
        <Modal id="control-panel-modal" title="Панель управления">
          <div class="control-panel">
            <Tabs id="control-panel-tabs" itemIdPrefix="control-panel-tab" align="center" className="control-panel-tabs" selectedValue={props.state().selectedControlPanelTab} tabs={controlPanelTabs} />
            <Switch>
              <Match when={props.state().selectedControlPanelTab === "object"}>
                <Show when={props.state().selfObject} fallback={<div class="control-panel-empty-page" />}>
                  {(_selfObject) => (
                    <div class="control-panel-object-page">
                      <div class="control-panel-object-row">
                        <GameFormRowLabel className="control-panel-object-row__label">Название модели космического объекта</GameFormRowLabel>
                        <div class="control-panel-object-row__value control-panel-object-row__value--readonly">{modelTitle()}</div>
                      </div>
                      <div class="control-panel-object-row">
                        <GameFormRowLabel className="control-panel-object-row__label">Включен</GameFormRowLabel>
                        <div class="control-panel-object-row__value control-panel-object-row__value--control">
                          <Checkbox id="control-panel-object-enabled" label="" checked={props.state().controlPanelObjectEnabled} />
                        </div>
                      </div>
                      <div class="control-panel-object-row">
                        <GameFormRowLabel className="control-panel-object-row__label">Пользовательское название объекта</GameFormRowLabel>
                        <div class="control-panel-object-row__value control-panel-object-row__value--control">
                          <TextInput
                            id="control-panel-object-title-input"
                            text={props.state().controlPanelObjectTitleText}
                            selectionStart={props.state().controlPanelObjectTitleSelectionStart}
                            selectionEnd={props.state().controlPanelObjectTitleSelectionEnd}
                            focused={props.state().controlPanelObjectTitleFocused}
                          />
                        </div>
                      </div>
                      <For each={rows()}>
                        {(row) => (
                          <div class="control-panel-object-row">
                            <GameFormRowLabel className="control-panel-object-row__label">{row.label}</GameFormRowLabel>
                            <div class="control-panel-object-row__value control-panel-object-row__value--readonly">{row.value}</div>
                          </div>
                        )}
                      </For>
                    </div>
                  )}
                </Show>
              </Match>
              <Match when={props.state().selectedControlPanelTab === "equipment"}>
                <div class="control-panel-equipment-page">
                  <Tabs id="control-panel-equipment-tabs" itemIdPrefix="control-panel-equipment-tab" align="center" className="control-panel-equipment-tabs" selectedValue={props.state().selectedControlPanelEquipmentTab} tabs={controlPanelEquipmentTabs} />
                  <Switch>
                    <Match when={props.state().selectedControlPanelEquipmentTab === "setup"}>
                      <div class="control-panel-equipment-layout">
                        <div class="control-panel-equipment-list">
                          <ListBox
                            id="control-panel-equipment-list"
                            selectedValue={selectedEquipment() ? String(selectedEquipment()?.group.ID) : ""}
                            items={equipmentGroups().map((equipment) => ({ value: String(equipment.group.ID), label: equipment.modelTitle }))}
                            scrollOffsetPx={props.state().controlPanelEquipmentListScroll.contentOffsetPx}
                          />
                          <Show when={props.state().controlPanelEquipmentListScroll.visible}>
                            <Scrollbar
                              id="control-panel-equipment-list-scrollbar"
                              className="control-panel-equipment-list-scrollbar"
                              thumbTopPercent={props.state().controlPanelEquipmentListScroll.thumbTopPercent}
                              thumbHeightPercent={props.state().controlPanelEquipmentListScroll.thumbHeightPercent}
                              dragging={props.state().controlPanelEquipmentListScroll.dragging}
                            />
                          </Show>
                        </div>
                        <div class="control-panel-equipment-info">
                          <Show when={selectedEquipment()} fallback={<div class="control-panel-empty-page" />}>
                            {(equipment) => (
                              <>
                                <div class="control-panel-object-row">
                                  <GameFormRowLabel className="control-panel-object-row__label">Название модели оборудования</GameFormRowLabel>
                                  <div class="control-panel-object-row__value control-panel-object-row__value--readonly">{equipment().modelTitle}</div>
                                </div>
                                <div class="control-panel-object-row">
                                  <GameFormRowLabel className="control-panel-object-row__label">Включено</GameFormRowLabel>
                                  <div class="control-panel-object-row__value control-panel-object-row__value--control">
                                    <Checkbox id="control-panel-equipment-enabled" label="" checked={selectedEquipmentEnabled()} />
                                  </div>
                                </div>
                                <div class="control-panel-object-row">
                                  <GameFormRowLabel className="control-panel-object-row__label">Количество включенных единиц</GameFormRowLabel>
                                  <div class="control-panel-object-row__value control-panel-object-row__value--control">
                                    <Slider
                                      id="control-panel-equipment-enabled-slider"
                                      value={selectedEquipmentEnabledCount()}
                                      min={0}
                                      max={Math.max(1, equipment().group.Count)}
                                      label={`${formatMetric(selectedEquipmentEnabledCount())} / ${formatMetric(equipment().group.Count)}`}
                                    />
                                  </div>
                                </div>
                                <For each={selectedEquipmentRows()}>
                                  {(row) => (
                                    <div class="control-panel-object-row">
                                      <GameFormRowLabel className="control-panel-object-row__label">{row.label}</GameFormRowLabel>
                                      <div class="control-panel-object-row__value control-panel-object-row__value--readonly">{row.value}</div>
                                    </div>
                                  )}
                                </For>
                                <div class="control-panel-equipment-action">
                                  <Button id="control-panel-equipment-usage-button" label="Использование" />
                                </div>
                              </>
                            )}
                          </Show>
                        </div>
                      </div>
                    </Match>
                    <Match when={props.state().selectedControlPanelEquipmentTab === "usage"}>
                      <div class="control-panel-empty-page" />
                    </Match>
                  </Switch>
                </div>
              </Match>
              <Match when={props.state().selectedControlPanelTab === "pilotTools"}>
                <div class="control-panel-empty-page" />
              </Match>
              <Match when={props.state().selectedControlPanelTab === "schemas"}>
                <div class="control-panel-empty-page" />
              </Match>
              <Match when={props.state().selectedControlPanelTab === "blueprints"}>
                <div class="control-panel-empty-page" />
              </Match>
              <Match when={props.state().selectedControlPanelTab === "map"}>
                <div class="control-panel-empty-page" />
              </Match>
            </Switch>
          </div>
        </Modal>
      </GameWindowLayer>
    </Show>
  );
};

const getControlPanelObjectRows = (state: GameUiState): ControlPanelObjectRow[] => {
  const object = state.selfObject;
  if (!object) {
    return [];
  }

  return [
    { label: "Никнейм аккаунта персонажа-владельца", value: emptyValue() },
    { label: "Никнейм аккаунта персонажа-создателя", value: emptyValue() },
    { label: "Масса", value: formatMetric(object.Mass) },
    { label: "Объём оборудования / Вместимость", value: formatPair(object.OccupiedVolume, object.Capacity) },
    { label: "Броня / Максимум брони", value: formatPair(object.Armor, object.MaxArmor) },
    { label: "Сложность", value: formatMetric(object.Complexity) },
    { label: "Максимальная скорость", value: formatPreciseMetric(object.MaxSpeed) },
    { label: "Максимальная угловая скорость", value: formatPreciseMetric(object.MaxAngularSpeed) },
    { label: "Продольная сила тяги (максимальная)", value: formatPreciseMetric(object.MaxAlongForce) },
    { label: "Поперечная сила тяги (максимальная)", value: formatPreciseMetric(object.MaxAcrossForce) },
    { label: "Крутящий момент (максимальный)", value: formatPreciseMetric(object.MaxTorque) },
    { label: "Потребляемая мощность / Вырабатываемая мощность", value: formatPair(object.ConsumingPower, object.GeneratingPower) },
    { label: "Запас топлива / Максимальный запас топлива", value: formatPair(object.Fuel, object.MaxFuel) },
    { label: "Занято на складе / Объём склада", value: formatPair(object.OccupiedVolume, object.Capacity) },
  ];
};

const getControlPanelEquipmentGroups = (state: GameUiState): ControlPanelEquipmentView[] => {
  const objectId = state.selfObject?.ID;
  if (!objectId) {
    return [];
  }

  return state.equipmentGroups
    .filter((group) => group.CosmicObjectID === objectId)
    .sort((left, right) => left.ID - right.ID)
    .map((group) => {
      const itemModel = state.referenceData?.ItemModel.Items[String(group.EquipmentItemModelID)];
      return {
        group,
        modelTitle: getReferenceTitle(itemModel) ?? group.Title,
        itemModel,
      };
    });
};

const getSelectedControlPanelEquipment = (groups: ControlPanelEquipmentView[], selectedGroupId: number | null): ControlPanelEquipmentView | null =>
  groups.find((equipment) => equipment.group.ID === selectedGroupId) ?? groups[0] ?? null;

const getControlPanelEquipmentEnabled = (state: GameUiState, equipment: ControlPanelEquipmentView | null): boolean =>
  equipment ? state.controlPanelEquipmentEnabledDrafts[equipment.group.ID] ?? equipment.group.Enabled : false;

const getControlPanelEquipmentEnabledCount = (state: GameUiState, equipment: ControlPanelEquipmentView | null): number =>
  equipment ? clampNumber(state.controlPanelEquipmentEnabledCountDrafts[equipment.group.ID] ?? equipment.group.EnabledCount, 1, Math.max(1, equipment.group.Count)) : 1;

const getControlPanelEquipmentInfoRows = (equipment: ControlPanelEquipmentView | null): ControlPanelEquipmentInfoRow[] => {
  if (!equipment) {
    return [];
  }
  const group = equipment.group;
  const model = equipment.itemModel;

  return [
    { label: "Активно", value: yesNo(group.Active) },
    { label: "Масса", value: formatModelMetric(model, "Mass") },
    { label: "Объём", value: formatModelMetric(model, "Volume") },
    { label: "Потребляемая мощность", value: formatModelMetric(model, "ConsumingPower") },
    { label: "Вырабатываемая мощность", value: formatModelMetric(model, "GeneratingPower") },
    { label: "Продольная сила тяги", value: formatModelMetric(model, "MaxAlongForce") },
    { label: "Поперечная сила тяги", value: formatModelMetric(model, "MaxAcrossForce") },
    { label: "Крутящий момент", value: formatModelMetric(model, "MaxTorque") },
    { label: "Сложность", value: formatModelMetric(model, "Complexity") },
  ];
};

const getControlPanelModelTitle = (state: GameUiState): string => {
  const object = state.selfObject;
  const model = object ? state.referenceData?.CosmicObjectModel.Items[String(object.CosmicObjectModelID)] : undefined;
  return getReferenceTitle(model) ?? emptyValue();
};

const getReferenceTitle = (model: CosmicObjectModelReference | ItemModelReference | undefined): string | null => {
  if (!model) {
    return null;
  }
  return stringField(model, "TitleRu") ?? stringField(model, "TitleEn") ?? stringField(model, "Acronym");
};

const stringField = (record: Record<string, unknown>, key: string): string | null => {
  const value = record[key];
  return typeof value === "string" && value.trim() !== "" ? value : null;
};

const emptyValue = (): string => "—";
const formatMetric = (value: number): string => formatNumber(value, 0);
const formatPreciseMetric = (value: number): string => formatNumber(value, 2);
const formatPair = (current: number, maximum: number): string => `${formatMetric(current)} / ${formatMetric(maximum)}`;
const clampNumber = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
const yesNo = (value: boolean): string => value ? "Да" : "Нет";
const numericField = (record: Record<string, unknown> | undefined, key: string): number => {
  const value = record?.[key];
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
};
const formatModelMetric = (model: ItemModelReference | undefined, key: string): string => formatMetric(numericField(model, key));

// Показывает десять ячеек инструментов пилота в центральной нижней части экрана.
const PilotToolbarPanel = (props: PilotToolbarPanelProps) => (
  <Show when={getPilotToolbarReadyState(props.state())}>
    {(ready) => {
      const toolbar = () =>
        getPilotToolbarView({
          selfObject: ready().selfObject,
          equipmentGroups: props.state().equipmentGroups,
          referenceData: ready().referenceData,
          selectedToolIndex: props.state().selectedPilotToolIndex,
        });

      return (
        <HudPanel position="bottom-center" className="pilot-toolbar" ariaLabel="Панель инструментов пилота">
          <Show when={toolbar().magazine}>
            {(magazine) => (
              <div class="pilot-toolbar__magazine">
                <div class="pilot-toolbar__magazine-fill" style={{ width: `${magazine().fillPercent}%` }} />
                <div class="pilot-toolbar__magazine-value">{magazine().valueText}</div>
              </div>
            )}
          </Show>
          <div class="pilot-toolbar__slots">
            <For each={toolbar().slots}>{(slot) => <PilotToolSlot slot={slot} />}</For>
          </div>
        </HudPanel>
      );
    }}
  </Show>
);

// Рисует одну квадратную ячейку инструмента пилота.
const PilotToolSlot = (props: PilotToolSlotProps) => (
  <div class={`pilot-tool-slot ${props.slot.isSelected ? "is-selected" : ""}`} title={props.slot.tool?.title ?? ""}>
    <Show when={props.slot.tool} fallback={<span class="pilot-tool-slot__key">{props.slot.index}</span>}>
      {(tool) => (
        <>
          <span class="pilot-tool-slot__key">{props.slot.index}</span>
          <Show when={tool().iconFilePath} fallback={<span class="pilot-tool-slot__fallback-icon">{tool().acronym.slice(0, 2)}</span>}>
            {(iconFilePath) => <img class="pilot-tool-slot__icon" src={`/assets/ui/icons/${iconFilePath().replace(/^assets\/ui\/icons\//, "")}`} alt="" />}
          </Show>
          <span class="pilot-tool-slot__count">{tool().enabledCount}</span>
        </>
      )}
    </Show>
  </div>
);

type MinimapReadyState = {
  // Посещаемый объект игрока, относительно которого центрируется карта.
  selfObject: NonNullable<GameUiState["selfObject"]>;
  // Справочники клиента, нужные для определения типов объектов.
  referenceData: NonNullable<GameUiState["referenceData"]>;
};

type MinimapPointProps = {
  // Данные точки объекта на мини-карте.
  point: MinimapPointView;
};

// Рисует статус стояночного якоря морским якорем.
const AnchorIcon = () => (
  <svg class="minimap-status__anchor-icon" viewBox="0 0 24 24" aria-hidden="true">
    <circle data-icon-part="anchor-ring" cx="12" cy="5" r="2.2" />
    <path data-icon-part="anchor-stock" d="M12 7.2v11.3M8 10h8" />
    <path data-icon-part="anchor-flukes" d="M5.5 15.5c1.1 3.1 3.5 4.8 6.5 4.8s5.4-1.7 6.5-4.8M5.5 15.5H8M18.5 15.5H16" />
    <path data-icon-part="anchor-tip" d="M12 18.5l-1.7-1.7M12 18.5l1.7-1.7" />
  </svg>
);

// Возвращает готовые данные для мини-карты только после получения объекта и справочников.
const getMinimapReadyState = (state: GameUiState): MinimapReadyState | null => {
  if (!state.selfObject || !state.referenceData) {
    return null;
  }
  return {
    selfObject: state.selfObject,
    referenceData: state.referenceData,
  };
};

// Показывает мини-карту в правой нижней части экрана.
const MinimapPanel = (props: MinimapPanelProps) => (
  <Show when={getMinimapReadyState(props.state())}>
    {(ready) => {
      const minimap = () =>
        getMinimapView({
          selfObject: ready().selfObject,
          objects: props.state().objects,
          referenceData: ready().referenceData,
        });

      return (
        <HudPanel position="right-bottom" className="minimap" ariaLabel="Мини-карта">
          <div class="minimap-compass">
            <For each={minimap().compassMarks}>
              {(mark) => (
                <span class="minimap-compass__mark" style={{ left: `${mark.xPercent}%` }}>
                  {mark.label}
                </span>
              )}
            </For>
          </div>
          <div class="minimap-body">
            <div class="minimap-status">
              <div class={`minimap-status__item minimap-status__zone ${minimap().isPveZone ? "is-pve" : "is-pvp"}`} title={minimap().isPveZone ? "ПВЕ" : "ПВП"}>
                {minimap().isPveZone ? "PVE" : "PVP"}
              </div>
              <div class={`minimap-status__item minimap-status__anchor ${minimap().isAnchored ? "is-active" : ""}`} title="Якорь">
                <AnchorIcon />
              </div>
            </div>
            <div class="minimap-map">
              <For each={minimap().points}>{(point) => <MinimapPoint point={point} />}</For>
              <div class="minimap-map__crosshair" />
            </div>
          </div>
        </HudPanel>
      );
    }}
  </Show>
);

// Рисует одну точку объекта внутри области мини-карты.
const MinimapPoint = (props: MinimapPointProps) => (
  <span
    class={`minimap-point minimap-point--${props.point.kind} ${props.point.isSelf ? "is-self" : ""}`}
    style={{ left: `${props.point.xPercent}%`, top: `${props.point.yPercent}%` }}
  />
);

type DebugOverlayProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

// Показывает диагностические строки поверх canvas.
const DebugOverlay = (props: DebugOverlayProps) => (
  <HudPanel id="debug-overlay" position="left-top" className="debug-overlay">
    {getDebugOverlayLines({
      status: props.state().status,
      selfObject: props.state().selfObject,
      textureFilePath: props.state().textureFilePath,
      fps: props.state().fps,
      zoom: props.state().zoom,
    }).join("\n")}
  </HudPanel>
);
