---
id: cairn:pilotSyncExpectations
schema_version: 1
title: Pilot Sync Expectations
slug: pilot-sync-expectations
type: investigation
status: working
created: 2026-05-07T12:00:00Z
updated: 2026-05-07T12:00:00Z
authors: [cairn]
actors: [codex]
source: pilot-fixture
tags: [pilot, sync]
---

# Pilot Sync Expectations

Cairn sync is intentionally conservative. If local and remote copies both change after the last shared base, Cairn reports a conflict and refuses the mutation.
