import { createEffect, createMemo, createSignal, For, Match, onCleanup, onMount, Show, Switch, type Accessor, type JSX } from "solid-js";
import { Portal } from "solid-js/web";
import type { GameUiState, SettingsTabValue } from "./gameUiState";
import { getDebugOverlayLines } from "./debugOverlay";
import { getObjectIndicators, type ObjectIndicatorView } from "./objectIndicators";
import { getMinimapView, type MinimapPointView } from "./minimap";
import { getPilotToolbarView, type PilotToolSlotView } from "./pilotToolbar";
import { getInformationPanelView, type InformationPanelRow } from "./informationPanel";
import { getInputEventOptions, getInputSettingsRows } from "./inputSettings";
import { Button, Checkbox, ContextMenu, Dropdown, EditControl, ListBox, Modal, NumericStepper, RadioGroup, Scrollbar, Slider, Splitter, Tabs, Tooltip, TreeView, VirtualList } from "../ui-kit/components";

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

type GameWindowLayerProps = {
  // Вид окна, который задаёт необходимые отличия поверх общего каркаса.
  variant: "settings" | "showcase";
  // Содержимое окна в общем экранном слое.
  children: JSX.Element;
};

const settingsTabs: Array<{ value: SettingsTabValue; label: string }> = [
  { value: "video", label: "Видео" },
  { value: "audio", label: "Аудио" },
  { value: "input", label: "Ввод" },
];

// Задаёт общий экранный шаблон для всех игровых модальных окон.
const GameWindowLayer = (props: GameWindowLayerProps) => (
  <div class={`game-window-layer game-window-layer--${props.variant}`}>
    {props.children}
  </div>
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
  let chatInputViewport: HTMLDivElement | undefined;
  let chatInputTextMeasure: HTMLSpanElement | undefined;
  let chatInputCaretMeasure: HTMLSpanElement | undefined;
  const [chatInputMetrics, setChatInputMetrics] = createSignal({ textOffsetPx: 0, caretLeftPx: 0 });

  // Рассчитывает горизонтальный сдвиг так, чтобы каретка оставалась внутри видимой части строки.
  const updateChatInputMetrics = () => {
    const viewportWidth = chatInputViewport?.getBoundingClientRect().width ?? 0;
    const textWidth = chatInputTextMeasure?.getBoundingClientRect().width ?? 0;
    const caretWidth = chatInputCaretMeasure?.getBoundingClientRect().width ?? 0;
    if (viewportWidth <= 0) {
      setChatInputMetrics({ textOffsetPx: 0, caretLeftPx: caretWidth });
      return;
    }

    const edgePaddingPx = Math.min(12, viewportWidth * 0.08);
    const maxOffsetPx = Math.max(0, Math.max(textWidth, caretWidth) - viewportWidth + edgePaddingPx);
    const previousOffsetPx = chatInputMetrics().textOffsetPx;
    let nextOffsetPx = previousOffsetPx;
    if (caretWidth - nextOffsetPx > viewportWidth - edgePaddingPx) {
      nextOffsetPx = caretWidth - viewportWidth + edgePaddingPx;
    }
    if (caretWidth - nextOffsetPx < edgePaddingPx) {
      nextOffsetPx = caretWidth - edgePaddingPx;
    }

    const textOffsetPx = Math.max(0, Math.min(maxOffsetPx, nextOffsetPx));
    setChatInputMetrics({
      textOffsetPx,
      caretLeftPx: Math.max(0, Math.min(viewportWidth, caretWidth - textOffsetPx)),
    });
  };

  createEffect(() => {
    props.state().chatInputText;
    props.state().chatCursorIndex;
    queueMicrotask(updateChatInputMetrics);
  });

  onMount(() => {
    window.addEventListener("resize", updateChatInputMetrics);
  });

  onCleanup(() => {
    window.removeEventListener("resize", updateChatInputMetrics);
  });

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
          <Tabs
            id="chat-tabs"
            itemIdPrefix="chat-tab"
            className="chat-tabs"
            itemClassName="chat-tab"
            selectedValue={String(tab().chatId)}
            tabs={chatTabs()}
          />
          <Show when={props.state().chatError}>
            {(error) => <div class="chat-error" style={{ "animation-name": chatErrorAnimationName() }}>{error()}</div>}
          </Show>
          <div id="chat-input" data-ui-kind="edit" class={`ui-kit-control ui-kit-edit chat-input ${props.state().chatInputFocused ? "is-focused" : ""}`}>
            <div class="chat-input__viewport" ref={chatInputViewport}>
              <span
                class="chat-input__text"
                style={{ transform: `translateX(${-chatInputMetrics().textOffsetPx}px)` }}
              >
                <span>{props.state().chatInputText.slice(0, chatSelectionStart())}</span>
                <Show when={chatSelectionEnd() > chatSelectionStart()}>
                  <span class="chat-input__selection">{props.state().chatInputText.slice(chatSelectionStart(), chatSelectionEnd())}</span>
                </Show>
                <span>{props.state().chatInputText.slice(chatSelectionEnd())}</span>
              </span>
              <span class="chat-input__measure" ref={chatInputTextMeasure} aria-hidden="true">{props.state().chatInputText}</span>
              <span class="chat-input__measure" ref={chatInputCaretMeasure} aria-hidden="true">{props.state().chatInputText.slice(0, props.state().chatCursorIndex)}</span>
              <Show when={props.state().chatInputFocused}>
                <span class="chat-input__caret" style={{ left: `${chatInputMetrics().caretLeftPx}px` }} />
              </Show>
            </div>
          </div>
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
  const options = createMemo(() => getInputEventOptions(props.state().referenceData));
  return (
    <Show when={props.state().settingsVisible}>
      <GameWindowLayer variant="settings">
        <Modal id="settings-modal" title="Настройки">
          <div class="settings-modal">
            <Tabs id="settings-tabs" itemIdPrefix="settings-tab" className="settings-tabs" selectedValue={props.state().selectedSettingsTab} tabs={settingsTabs} />
            <Switch>
              <Match when={props.state().selectedSettingsTab === "video"}>
                <div class="settings-empty-page" />
              </Match>
              <Match when={props.state().selectedSettingsTab === "audio"}>
                <div class="settings-empty-page" />
              </Match>
              <Match when={props.state().selectedSettingsTab === "input"}>
                <div class="settings-input-table">
                  <div class="settings-input-table__content" style={{ transform: `translateY(-${props.state().inputSettingsScroll.contentOffsetPx}px)` }}>
                    <For each={rows()}>
                      {(row) => (
                        <div class="settings-input-row">
                          <div class="settings-input-row__action">{row.actionTitle}</div>
                          <Dropdown
                            id={`settings-input-select-${row.actionTypeId}`}
                            selectedValue={String(row.inputEventTypeId)}
                            open={props.state().openInputSettingsActionId === row.actionTypeId}
                            options={options()}
                            menuScroll={props.state().inputSettingsDropdownScroll}
                          />
                        </div>
                      )}
                    </For>
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
              <Show when={props.state().inputSettingsError}>
                {(error) => <div class="settings-modal__error">{error()}</div>}
              </Show>
              <Button id="settings-cancel-button" label="Отмена" />
              <Button id="settings-save-button" label={props.state().inputSettingsSaving ? "Сохранение" : "Сохранить"} />
            </div>
          </div>
        </Modal>
      </GameWindowLayer>
    </Show>
  );
};

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
