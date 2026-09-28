# medical-service

Pharmacy and drug inventory (nhà thuốc / kho dược): medicine catalogue, stock by batch with expiry dates,
dispensing with FEFO and drug-safety checks. Patients and prescriptions live in
`hospital-patient-management-service`; this service refers to them only by `patient_ref` / `prescription_ref`.

Hexagonal architecture wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                            fx.New(infrastructure.Module)
internal/domain/model                  entities and pure rules: FEFO allocation, safety evaluation, Date
internal/domain/port                   driving (use case) and driven (repository, tx, clock) interfaces
internal/application                   use cases: medicines, inventory, interactions, dispensing
internal/adapter/primary/http          gin handlers and routes
internal/adapter/secondary/repository  MySQL (gorm), transaction carried in context
internal/infrastructure                config, clock, database, HTTP server, fx module
```

## Business rules

- **Expiry**: a batch is never dispensed on or after its expiry date. "Today" is the calendar day in
  `PHARMACY_TIMEZONE`. Already-expired batches cannot be received.
- **FEFO**: dispensing takes stock from the batch that expires first. All items of a dispense succeed
  or none do (one transaction, batch rows locked `FOR UPDATE`, in medicine-id order to avoid deadlocks),
  so concurrent dispenses cannot oversell.
- **Prescription**: medicines with `requires_prescription` need a `prescription_ref`.
- **Safety check**, run before every dispense against the medicines being dispensed and the patient's
  `current_medicine_ids` and `allergies`:

  | Warning | Severity | Effect |
  |---|---|---|
  | Allergy: an allergy matches an ingredient or `allergen_groups` (e.g. `penicillin` for amoxicillin) | CONTRAINDICATED | blocked, cannot be overridden |
  | Interaction rule with severity CONTRAINDICATED | CONTRAINDICATED | blocked |
  | Interaction rule with severity MAJOR | MAJOR | needs `override_reason` |
  | Two medicines with the same active ingredient (e.g. two paracetamol products) | MAJOR | needs `override_reason` |
  | Interaction rule MINOR / MODERATE | as set | recorded on the dispense only |

  Names are compared after resolving synonyms, and an allergy also matches the reference allergen groups
  of each ingredient, so a medicine created without `allergen_groups` is still caught (see Terminology).
  Matching is by drug class, not cross-class: an allergy to `penicillin` does not block cephalosporins;
  record `beta-lactam` for patients who must avoid the whole family.
- **Terminology** (`ingredient_aliases`, `ingredient_allergen_groups`):
  - Aliases map synonyms to one canonical name, e.g. `aspirin` / `acid acetylsalicylic` → `acetylsalicylic acid`,
    `acetaminophen` → `paracetamol`, `cephalexin` → `cefalexin`. Aliases are one hop deep; the API rejects chains.
  - Allergen groups map ingredients to drug classes: penicillin, cephalosporin, carbapenem (all also
    beta-lactam), sulfonamide (antibiotics), nsaid, macrolide, fluoroquinolone, tetracycline, opioid.
  - Ingredients are stored canonically when a medicine is created, and resolved again at every check.
- **Reference data**: `database.sql` seeds 13 aliases, 59 ingredient-group links and 26 well-documented
  interactions (warfarin with NSAIDs / fluconazole / metronidazole, simvastatin with strong CYP3A4
  inhibitors, PDE5 inhibitors with nitrates, ...) with `source = reference`. They are inserted with
  `INSERT IGNORE` on every start, so anything a pharmacist adds or edits through the API (`source = local`)
  is never overwritten. The list is a starting point, not a complete drug-interaction database: a
  pharmacist should review it and extend it for the formulary in use.
- **Stock ledger**: every change (RECEIVE, DISPENSE, ADJUST, DISPOSE) is a signed row in `stock_movements`.

## API

| Method | Path | |
|---|---|---|
| GET | `/health` | |
| POST | `/api/v1/medicines` | `code`, `name`, `ingredients`, `allergen_groups`, `dosage_form`, `strength`, `unit`, `requires_prescription`, `reorder_level` |
| GET | `/api/v1/medicines?q=&limit=&offset=` | search by name, code or ingredient |
| GET | `/api/v1/medicines/:id` | |
| GET | `/api/v1/medicines/:id/stock` | available and expired units, batches |
| GET | `/api/v1/medicines/:id/movements` | stock ledger, newest first |
| POST | `/api/v1/medicines/:id/batches` | receive stock: `batch_number`, `expiry_date` (YYYY-MM-DD), `quantity`, `unit_cost`, `supplier`, `received_by`. Same batch number + same expiry adds to it |
| POST | `/api/v1/batches/:id/adjust` | stock count: `quantity` (counted), `reason`, `adjusted_by` |
| GET | `/api/v1/inventory/expiring?days=30` | batches with stock expiring within N days |
| GET | `/api/v1/inventory/low-stock` | unexpired stock at or below `reorder_level` |
| POST | `/api/v1/inventory/dispose-expired` | write off all expired stock: `disposed_by` |
| POST | `/api/v1/interactions` | `ingredient_a`, `ingredient_b`, `severity`, `description`; same pair updates the rule |
| GET | `/api/v1/interactions` | |
| POST | `/api/v1/interactions/check` | `medicine_ids`, `current_medicine_ids`, `allergies` → warnings |
| POST | `/api/v1/dispenses` | `patient_ref`, `prescription_ref`, `dispensed_by`, `items[{medicine_id, quantity}]`, `current_medicine_ids`, `allergies`, `override_reason` |
| GET | `/api/v1/terminology` | aliases and allergen groups |
| POST | `/api/v1/terminology/aliases` | `alias`, `canonical` |
| POST | `/api/v1/terminology/allergen-groups` | `ingredient`, `allergen_group` |
| GET | `/api/v1/dispenses?patient_ref=` | dispensing history of a patient |
| GET | `/api/v1/dispenses/:id` | |

Errors: 400 invalid input, 404 not found, 409 conflict / insufficient stock / safety check (body includes
`warnings`), 422 prescription required.

## Configuration

| Env | Default |
|---|---|
| `PORT_HTTP_SERVER` | `8093` |
| `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_DBNAME` | `localhost` / `3306` / `root` / empty / `medical_service` |
| `PHARMACY_TIMEZONE` | `Asia/Ho_Chi_Minh` |

The schema in `database.sql` is applied at startup. Setting `HostConsul`, `KeyConsul` and `ServiceConsul`
in `.env` loads the same config from Consul instead.

```
make run
make test
```
