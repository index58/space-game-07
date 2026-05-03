import { For, Match, Show, Switch, type Accessor, type JSX } from "solid-js";
import type { GameUiState } from "./gameUiState";
import { getDebugOverlayLines } from "./debugOverlay";
import { getObjectIndicators, type ObjectIndicatorView } from "./objectIndicators";
import { getMinimapView, type MinimapPointView } from "./minimap";
import { getPilotToolbarView, type PilotToolSlotView } from "./pilotToolbar";

type HudPanelPosition = "left-bottom" | "bottom-center" | "right-bottom" | "left-top";

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
    <PilotToolbarPanel state={props.state} />
    <MinimapPanel state={props.state} />
    <DebugOverlay state={props.state} />
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
              <div class={`minimap-status__zone ${minimap().isPveZone ? "is-pve" : "is-pvp"}`} title={minimap().isPveZone ? "ПВЕ" : "ПВП"} />
              <div class={`minimap-status__anchor ${minimap().isAnchored ? "is-active" : ""}`} title="Якорь" />
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
