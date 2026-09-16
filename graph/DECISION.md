# Graph decision

## Decision
Keep `graph/` for now, but stop treating Go graph rendering/organization logic as the product surface.

The canonical graph representation is:

- `organization.json`: machine-readable state
- `organization.mmd`: human-readable visualization

The Go graph implementation is transitional and may be removed after the AW-driven organization state is proven.

## TUI decision surface

`soubi.tui` / `aw.tui` should not decide *how* repositories are organized. It should expose a small set of events that AW can execute:

1. `refresh` — refresh recent repositories
2. `inspect` — inspect current organization state
3. `plan` — ask the coordinator to propose work
4. `run` — execute the selected AW
5. `approve` — approve a proposed plan

The TUI is therefore a **control surface for AW**, not an orchestration engine.

## First executable workflow

Start with one safe, observable workflow:

`TUI RUN → AW → latest 30 repos → repo.yaml discovery → organization.json/mmd → TUI refresh`

Once this loop works, add BQML and pi-agent delegation.
