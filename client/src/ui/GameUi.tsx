import { For, Match, Show, Switch, type Accessor, type JSX } from "solid-js";
import type { GameUiState } from "./gameUiState";
import { getDebugOverlayLines } from "./debugOverlay";
import { getObjectIndicators, type ObjectIndicatorView } from "./objectIndicators";
import { getMinimapView, type MinimapPointView } from "./minimap";
import { getPilotToolbarView, type PilotToolSlotView } from "./pilotToolbar";

type HudPanelPosition = "left-bottom" | "left-middle" | "bottom-center" | "right-bottom" | "left-top";

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
    <PilotToolbarPanel state={props.state} />
    <MinimapPanel state={props.state} />
    <DebugOverlay state={props.state} />
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

type PilotToolbarPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type ChatPanelProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type GameCursorProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

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

// Показывает доступные вкладки, последние строки истории и локальную строку ввода.
const ChatPanel = (props: ChatPanelProps) => {
  const selectedTab = () => props.state().chatState?.tabs.find((tab) => tab.chatId === props.state().chatState?.selectedChatId) ?? null;
  const chatCaretLeft = () => `calc(0.8vh + ${props.state().chatCursorIndex}ch)`;
  const chatErrorAnimationName = () => props.state().chatErrorSeq % 2 === 0 ? "chat-error-fade-even" : "chat-error-fade-odd";

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
              <div class={`chat-scrollbar ${props.state().chatScroll.dragging ? "is-dragging" : ""}`}>
                <div
                  class="chat-scrollbar__thumb"
                  style={{
                    top: `${props.state().chatScroll.thumbTopPercent}%`,
                    height: `${props.state().chatScroll.thumbHeightPercent}%`,
                  }}
                />
              </div>
            </Show>
          </div>
          <div class="chat-tabs">
            <For each={props.state().chatState?.tabs ?? []}>
              {(chatTab) => (
                <div class={`chat-tab ${chatTab.chatId === tab().chatId ? "is-selected" : ""}`}>
                  <span class="chat-tab__marker">{chatTab.communityTypeAcronym === "Server" ? "S" : "D"}</span>
                  <span class="chat-tab__title">{chatTab.title}</span>
                  <Show when={(chatTab.unreadCount ?? 0) > 0}>
                    <span class="chat-tab__unread">{chatTab.unreadCount}</span>
                  </Show>
                </div>
              )}
            </For>
          </div>
          <Show when={props.state().chatError}>
            {(error) => <div class="chat-error" style={{ "animation-name": chatErrorAnimationName() }}>{error()}</div>}
          </Show>
          <div class={`chat-input ${props.state().chatInputFocused ? "is-focused" : ""}`}>
            <span class="chat-input__text">{props.state().chatInputText}</span>
            <Show when={props.state().chatInputFocused}>
              <span class="chat-input__caret" style={{ left: chatCaretLeft() }} />
            </Show>
          </div>
          <Show when={props.state().chatContextMenu}>
            {(menu) => (
              <div
                class="chat-context-menu"
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
    <div class="game-cursor" style={{ left: `${props.state().gameCursor.x}px`, top: `${props.state().gameCursor.y}px` }} />
  </Show>
);

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
