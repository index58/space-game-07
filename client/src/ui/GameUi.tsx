import { For, Match, Show, Switch, type Accessor } from "solid-js";
import type { GameUiState } from "./gameUiState";
import { getDebugOverlayLines } from "./debugOverlay";
import { getObjectIndicators, type ObjectIndicatorView } from "./objectIndicators";

type GameUiProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

type ObjectIndicatorProps = {
  // Данные одной строки панели основных показателей.
  indicator: ObjectIndicatorView;
};

// Корневой компонент всех текущих UI-слоёв поверх Phaser canvas.
export const GameUi = (props: GameUiProps) => (
  <>
    <ObjectIndicatorsPanel selfObject={props.state().selfObject} />
    <DebugOverlay state={props.state} />
  </>
);

type ObjectIndicatorsPanelProps = {
  // Посещаемый объект игрока, если он уже получен.
  selfObject: GameUiState["selfObject"];
};

// Показывает основные показатели посещаемого объекта в левой нижней части экрана.
const ObjectIndicatorsPanel = (props: ObjectIndicatorsPanelProps) => (
  <Show when={props.selfObject}>
    {(selfObject) => (
      <section class="object-indicators-overlay" aria-label="Основные показатели посещаемого объекта">
        <For each={getObjectIndicators(selfObject())}>
          {(indicator) => <ObjectIndicator indicator={indicator} />}
        </For>
      </section>
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
        <path d="M9 6h4" />
        <path d="M15 8h2l2 3v7c0 1-1 2-2 2h-2" />
        <path d="M19 11h-2" />
      </Match>
    </Switch>
  </svg>
);

type DebugOverlayProps = {
  // Реактивное состояние всего игрового UI.
  state: Accessor<GameUiState>;
};

// Показывает диагностические строки поверх canvas.
const DebugOverlay = (props: DebugOverlayProps) => (
  <div id="debug-overlay">
    {getDebugOverlayLines({
      status: props.state().status,
      selfObject: props.state().selfObject,
      textureFilePath: props.state().textureFilePath,
      fps: props.state().fps,
      zoom: props.state().zoom,
    }).join("\n")}
  </div>
);
