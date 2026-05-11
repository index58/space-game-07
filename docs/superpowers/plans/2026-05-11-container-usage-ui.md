# Container Usage UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the "Контейнер" UI in `Панель управления -> Оборудование -> Использование` with immediate server-side movement between two selected containers.

**Architecture:** The server snapshot will include item groups so the client can render container contents. The client will keep separate left and right usage selections, render filtered selectors and container content blocks, and send a dedicated move command for all selected source contents.

**Tech Stack:** Go WebSocket protocol and world state, TypeScript Solid UI, Phaser scene action routing, Vitest, Go tests.

---

### Task 1: Server Contract And Move Logic

**Files:**
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/ws/protocol.go`
- Modify: `server/internal/world/world.go`
- Test: `server/internal/ws/protocol_test.go`
- Test: `server/internal/world/world_test.go`

- [ ] Write Go tests for snapshot `itemGroups`, decoding `controlPanelContainerTransfer`, successful movement, and invalid container rejection.
- [ ] Run targeted Go tests with elevated permission and verify they fail because the fields and command do not exist.
- [ ] Add `ItemGroups` to snapshots and encode it as `itemGroups`.
- [ ] Add the transfer WebSocket message and decoder.
- [ ] Add world logic that verifies both containers belong to the controlled object, both are container equipment, then moves all item groups from source to target, merging matching item models.
- [ ] Run targeted Go tests with elevated permission and verify they pass.

### Task 2: Client Protocol And UI State

**Files:**
- Modify: `client/src/network/protocol.ts`
- Modify: `client/src/network/GameClient.ts`
- Modify: `client/src/ui/gameUiState.ts`
- Modify: `client/src/game/GameScene.ts`
- Test: `client/src/network/GameClient.test.ts`

- [ ] Write TypeScript tests for sending `controlPanelContainerTransfer`.
- [ ] Run targeted Vitest and verify it fails because the client method does not exist.
- [ ] Add `ItemGroup`, snapshot `itemGroups`, and transfer message types.
- [ ] Add `sendControlPanelContainerTransfer`.
- [ ] Add left and right usage selection fields to UI state and feed `itemGroups` from snapshots.
- [ ] Route clicks for left selector, right selector, and transfer buttons.
- [ ] Run targeted Vitest and verify it passes.

### Task 3: Container Usage Rendering

**Files:**
- Modify: `client/src/ui/GameUi.tsx`
- Modify: `client/src/style.css`
- Test: `client/src/ui/GameUi.test.tsx`
- Test: `client/src/ui/style.test.ts`

- [ ] Write Vitest tests for the usage tab showing left containers, right internal equipment, two content blocks, and transfer buttons for a right container.
- [ ] Run targeted Vitest and verify it fails because the usage tab is still empty.
- [ ] Render two independent usage panels with filtered selectors.
- [ ] Render the "Содержимое контейнера" list with item model title and count columns.
- [ ] Render the right-side container subtype with two narrow icon-only transfer buttons.
- [ ] Add CSS for stable panel sizing, content lists, and auxiliary buttons.
- [ ] Run targeted Vitest and verify it passes.

### Task 4: Verification

**Files:**
- No production edits expected.

- [ ] Run client tests touched by the change.
- [ ] Run server tests touched by the change with elevated permission.
- [ ] Do not start the client or server.
- [ ] Report that visual browser verification was not run because project instructions forbid starting the client without explicit permission.
