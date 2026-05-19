# test-coverage.md — dev.helix.agent/skillregistry

Round 282 symbol → test / Challenge ledger. Every exported symbol of
`dev.helix.agent/skillregistry` (package name `agents`) MUST appear
here with the test(s) and Challenge(s) that exercise it AND the
anti-bluff dimension each proves. Adding an exported symbol without
updating this ledger is a CONST-048 violation. Per Article XI §11.9,
every PASS row MUST carry positive runtime evidence — the "Evidence"
column documents what to capture during a release-gate sweep.

## Exported symbols — types

| Symbol                            | Kind   | Unit test(s)                                  | Challenge(s)                              | Anti-bluff dimension                                                | Evidence (runtime)                              |
|-----------------------------------|--------|-----------------------------------------------|-------------------------------------------|---------------------------------------------------------------------|-------------------------------------------------|
| `Skill`                           | struct | `TestSkillManager_*`, `TestLoader_*`          | `runner -all`                             | YAML → struct round-trip is real, not regex.                        | Runner `loaded_skills=5` line.                  |
| `SkillDefinition`                 | struct | `TestSkillExecutor_*`                         | `runner -all`                             | Handler+timeout wiring is observable post-Register.                 | Runner asserts Handler value after Get.         |
| `SkillParameter`                  | struct | `TestSkillValidator_*`                        | n/a                                       | Parameter type-validation rejects unknown types.                    | Unit-test assertion.                            |
| `SkillReturn`                     | struct | `TestSkillValidator_*`                        | n/a                                       | Field present on round-trip.                                        | Unit-test assertion.                            |
| `SkillExample`                    | struct | `TestSkillValidator_*`                        | n/a                                       | Field present on round-trip.                                        | Unit-test assertion.                            |
| `SkillExecutionContext`           | struct | `TestSkillExecutor_*`                         | `runner -all`                             | Inputs/SessionID/UserID actually reach handler.                     | Handler captures Inputs len; runner prints it.  |
| `SkillResult`                     | struct | `TestSkillExecutor_*`                         | `runner -all`                             | Status/Output/Duration set by Success path.                         | Runner per-skill line shows status=success.     |
| `SkillStatus`                     | type   | `TestSkillManager_Enable*`                    | `runner -all`                             | Active vs Inactive transitions via Enable/Disable.                  | Runner asserts IsActive() post-Enable.          |
| `SkillCategory`                   | type   | `TestSkillManager_ListByCategory`             | `runner -all`                             | ListByCategory surfaces fixtures with matching category.            | Runner asserts general slice non-empty.         |
| `ExecutionStatus`                 | type   | `TestSkillExecutor_*`                         | `runner -all`                             | Success constant matched against real handler result.               | Runner asserts res.Status == Success.           |
| `SkillFilter`                     | struct | `TestSkillManager_Filter*`                    | `runner -all`                             | Enabled-only filter actually filters.                               | Runner asserts Filter(Enabled=true) len.        |
| `SkillMetrics`                    | struct | n/a (gap — see "Known gaps" in CLAUDE.md)     | `runner -all`                             | Metrics struct allocated at Register; SkillID matches.              | Runner prints metrics_id and metrics_total.     |
| `SkillStorage`                    | iface  | `TestInMemoryStorage_*`, `TestPostgresStorage_*` | `runner -all`                          | List round-trip on real backing store.                              | Runner reads storage.List after Register.       |
| `SkillHandler`                    | type   | `TestSkillExecutor_RegisterHandler`           | `runner -all`                             | Real func registered + dispatched.                                  | Runner registers challenge.exerciser.           |
| `ExecutionHook`                   | type   | `TestSkillExecutor_Hooks`                     | n/a                                       | Pre/post hook ordering preserved.                                   | Unit-test assertion.                            |
| `StorageConfig`                   | struct | `TestStorage_DefaultConfig`                   | n/a                                       | Default config has valid in-memory type.                            | Unit-test assertion.                            |
| `MemoryStorage`                   | struct | `TestInMemoryStorage_*`                       | `runner -all`                             | Save/Get/Delete/List round-trip.                                    | Runner exercises Register→List path.            |
| `PostgresStorage`                 | struct | `TestPostgresStorage_*`                       | n/a (integration tier)                    | SQL schema init + JSONB round-trip (when DB available).             | Integration-tier assertion (skipped w/o DB).    |
| `Loader`                          | struct | `TestLoader_LoadSkillFromFile`                | `runner -all`                             | Real file I/O + YAML parse, not regex.                              | Runner loads 5 real YAML fixtures.              |
| `SkillManager`                    | struct | `TestSkillManager_*`                          | `runner -all`                             | Composed Manager+Storage+Executor wiring is end-to-end.             | Runner exercises Register+Enable+Execute path.  |
| `SkillExecutor`                   | struct | `TestSkillExecutor_*`                         | `runner -all`                             | Concurrency semaphore + handler dispatch.                           | Runner Execute() returns Success.               |
| `SkillValidator`                  | struct | `TestSkillValidator_*`                        | `runner -all`, `describe-challenge --mutate` | Required-fields enforcement; mutation surfaces empty ID.        | Mutate-mode exit 99.                            |
| `DependencyResolver`              | struct | `TestSkillValidator_Dependencies*`            | n/a                                       | Cycle detection works on cooked graphs.                             | Unit-test assertion.                            |

## Exported symbols — constants

| Symbol                                | Anti-bluff dimension                                  | Where proved                                |
|---------------------------------------|-------------------------------------------------------|---------------------------------------------|
| `SkillStatusActive` / `…Inactive` / `…Disabled` / `…Error` | Enable/Disable produces expected Status transitions. | `runner -all` IsActive() check.             |
| `SkillCategoryCode` / `…Data` / `…DevOps` / `…Testing` / `…Security` / `…Monitoring` / `…General` | ListByCategory surfaces fixtures in the matching slot. | `runner -all` ListByCategory(general).      |
| `ExecutionStatusPending` / `…Running` / `…Success` / `…Failed` / `…Cancelled` / `…Timeout` | Success constant matched against real handler result. | `runner -all` Status assertion.             |
| `Err*` sentinels                      | Errors.Is path resolves correctly to sentinel.        | `TestSkillManager_*ErrorPaths`.             |

## Exported symbols — functions / constructors

| Symbol                              | Unit test(s)                          | Challenge(s)            | Anti-bluff dimension                          | Evidence                                   |
|-------------------------------------|---------------------------------------|-------------------------|-----------------------------------------------|--------------------------------------------|
| `NewSkillManager`                   | `TestSkillManager_*`                  | `runner -all`           | Returns wired Manager with storage+executor.  | Runner constructs + executes.              |
| `NewInMemoryStorage`                | `TestInMemoryStorage_*`               | `runner -all`           | Returns thread-safe store.                    | Runner storage.List round-trip.            |
| `NewMemoryStorage`                  | `TestInMemoryStorage_*`               | n/a                     | Config-based constructor.                     | Unit-test assertion.                       |
| `NewPostgresStorage`                | `TestPostgresStorage_*`               | n/a                     | SQL constructor + InitSchema.                 | Integration tier (skipped w/o DB).         |
| `NewStorage`                        | `TestStorage_NewStorage`              | n/a                     | Factory routes by Type.                       | Unit-test assertion.                       |
| `DefaultStorageConfig`              | `TestStorage_DefaultConfig`           | n/a                     | Default type=memory.                          | Unit-test assertion.                       |
| `NewLoader`                         | `TestLoader_*`                        | `runner -all`           | Returns loader with default formats.          | Runner loads 5 YAML fixtures.              |
| `NewSkillExecutor`                  | `TestSkillExecutor_*`                 | `runner -all`           | Default concurrency cap installed.            | Runner concurrent-dispatch works.          |
| `NewSkillExecutorWithConcurrency`   | `TestSkillExecutor_Concurrency`       | n/a                     | Custom cap applied.                           | Unit-test assertion.                       |
| `NewSkillValidator`                 | `TestSkillValidator_*`                | `runner -all`           | Returns ready-to-use validator.               | Runner ValidateSkill on every fixture.     |
| `NewDependencyResolver`             | `TestSkillValidator_Dependencies*`    | n/a                     | Returns resolver with empty visited/stack.    | Unit-test assertion.                       |
| `NewSkillExecutionContext`          | `TestSkillExecutor_*`                 | `runner -all`           | Generates unique ExecutionID per call.        | Runner per-skill execution.                |
| `NewSkillResult`                    | `TestSkillExecutor_*`                 | `runner -all`           | Starts in Pending; transitions via Success.   | Runner res.Status == Success.              |
| `CreateLoggingHook` / `…ValidationHook` | `TestSkillExecutor_Hooks`         | n/a                     | Returned hook callable.                       | Unit-test assertion.                       |

## Anti-bluff dimensions covered

| Dimension                                                           | Where proved                                                |
|---------------------------------------------------------------------|-------------------------------------------------------------|
| Real I/O (YAML files on disk, not embedded literal corpus)          | Runner `LoadSkillsFromDirectory("challenges/fixtures")`     |
| Real YAML parse (not regex)                                         | `yaml.Unmarshal` in `loader.loadSkillFromYAML`              |
| Real validator invariants enforced (not metadata-only)              | Runner `ValidateSkill` on every fixture                     |
| Real handler dispatch (not echo-stub)                               | Runner registers `challenge.exerciser` + asserts Success    |
| Real storage round-trip (not in-memory cache shadow)                | Runner `storage.List` post-Register                         |
| Discovery paths exercised (List/Filter/Search/ListByCategory)       | Runner asserts each surface returns expected count          |
| 5-locale bilingual UX (CONST-046)                                   | Runner prints `en/sr/ja/es/de` summary lines                |
| Paired-mutation evidence (CONST-035)                                | `--mutate` flag in `skillregistry_describe_challenge.sh`    |
| Honest gap-acknowledgement (no over-promising)                      | Metrics row tied to "Known gaps" in CLAUDE.md               |
| Skill enable/disable transition observable                          | Runner `IsActive()` after `Enable`                          |

## Maintenance

Every CL that touches `types.go`, `registry.go`, `manager.go`,
`executor.go`, `validator.go`, `loader.go`, or any `storage*.go`
file (adds/removes/renames an exported symbol, alters a struct
shape, changes loader / executor / validator behaviour) MUST update
this file in the SAME commit. Drift is a CONST-048 violation. The
Challenge runner exercises the full Register → Enable → Execute →
Metrics → Discovery loop at runtime — adding a new public method
without an accompanying assertion in the runner is a paired
CONST-035 violation.
