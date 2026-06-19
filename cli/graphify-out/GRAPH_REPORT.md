# Graph Report - .  (2026-06-07)

## Corpus Check
- 105 files · ~84,218 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 626 nodes · 868 edges · 75 communities (64 shown, 11 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 134 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Staging Removal|Staging Removal]]
- [[_COMMUNITY_Object Storage|Object Storage]]
- [[_COMMUNITY_Index & FileMode|Index & FileMode]]
- [[_COMMUNITY_Path Types & Workspace|Path Types & Workspace]]
- [[_COMMUNITY_Commit & Identity|Commit & Identity]]
- [[_COMMUNITY_Commit Service & Hash|Commit Service & Hash]]
- [[_COMMUNITY_Tree Resolver|Tree Resolver]]
- [[_COMMUNITY_Index Serialization|Index Serialization]]
- [[_COMMUNITY_Tree Operations|Tree Operations]]
- [[_COMMUNITY_Diff Engine|Diff Engine]]
- [[_COMMUNITY_Config & Validation|Config & Validation]]
- [[_COMMUNITY_Documentation Hub|Documentation Hub]]
- [[_COMMUNITY_Branch Switch|Branch Switch]]
- [[_COMMUNITY_Path Resolver|Path Resolver]]
- [[_COMMUNITY_Show Inspection|Show Inspection]]
- [[_COMMUNITY_Config Domain|Config Domain]]
- [[_COMMUNITY_CLI DiffShow|CLI Diff/Show]]
- [[_COMMUNITY_Reset Service|Reset Service]]
- [[_COMMUNITY_Ref Service|Ref Service]]
- [[_COMMUNITY_Restore Service|Restore Service]]
- [[_COMMUNITY_Branch Service|Branch Service]]
- [[_COMMUNITY_Commit Resolver|Commit Resolver]]
- [[_COMMUNITY_Status Inspection|Status Inspection]]
- [[_COMMUNITY_Myers Algorithm|Myers Algorithm]]
- [[_COMMUNITY_Design Rationales|Design Rationales]]
- [[_COMMUNITY_Init Service|Init Service]]
- [[_COMMUNITY_Add Service|Add Service]]
- [[_COMMUNITY_CLI Root Wiring|CLI Root Wiring]]
- [[_COMMUNITY_Update Ref|Update Ref]]
- [[_COMMUNITY_Update Index|Update Index]]
- [[_COMMUNITY_CatFile Inspection|CatFile Inspection]]
- [[_COMMUNITY_Change Detector|Change Detector]]
- [[_COMMUNITY_Error Handling|Error Handling]]
- [[_COMMUNITY_Symbolic Ref|Symbolic Ref]]
- [[_COMMUNITY_Config Storage|Config Storage]]
- [[_COMMUNITY_Index Storage|Index Storage]]
- [[_COMMUNITY_Commit Execution|Commit Execution]]
- [[_COMMUNITY_File Stat|File Stat]]
- [[_COMMUNITY_Defensive Patterns|Defensive Patterns]]
- [[_COMMUNITY_Staging Update|Staging Update]]
- [[_COMMUNITY_Documentation Patterns|Documentation Patterns]]
- [[_COMMUNITY_README|README]]
- [[_COMMUNITY_Refactoring Guide|Refactoring Guide]]
- [[_COMMUNITY_Layered Architecture|Layered Architecture]]
- [[_COMMUNITY_Config TOML Choice|Config TOML Choice]]

## God Nodes (most connected - your core abstractions)
1. `initializeServices()` - 35 edges
2. `Index` - 14 edges
3. `RemoveService` - 13 edges
4. `NewHashFromHex()` - 13 edges
5. `How to Write a Command Guide Document` - 10 edges
6. `ObjectService` - 9 edges
7. `RefService` - 9 edges
8. `TreeResolver` - 8 edges
9. `parseTreeEntries()` - 8 edges
10. `FileMode` - 8 edges

## Surprising Connections (you probably didn't know these)
- `initializeServices()` --calls--> `NewResetService()`  [INFERRED]
  cli/root.go → reset.go
- `initializeServices()` --calls--> `NewLogService()`  [INFERRED]
  cli/root.go → commit/log.go
- `initializeServices()` --calls--> `NewCommitService()`  [INFERRED]
  cli/root.go → commit/commit.go
- `initializeServices()` --calls--> `NewCommitTreeService()`  [INFERRED]
  cli/root.go → commit/commit_tree.go
- `initializeServices()` --calls--> `NewLsTreeService()`  [INFERRED]
  cli/root.go → tree/ls_tree.go

## Hyperedges (group relationships)
- **Index Staging and Tree Writing Workflow** — doc_add_command, doc_update_index_command, doc_write_tree_command, doc_ls_files_command, rationale_three_area_model, rationale_index_as_flat_list [INFERRED 0.85]
- **Error Handling Architecture (apperr Kind Pattern)** — doc_error_guide, doc_error_handling_guide, rationale_error_classification_kind, rationale_error_translation_boundary, rationale_cli_error_boundary [EXTRACTED 0.90]
- **Command Guides Following Template Specification** — doc_command_guide_template, doc_hash_object_command, doc_rm_command, doc_cat_file_command, doc_config_command, doc_add_command, doc_ls_files_command, doc_update_index_command, doc_init_command, doc_write_tree_command [EXTRACTED 1.00]

## Communities (75 total, 11 thin omitted)

### Community 0 - "Staging Removal"
Cohesion: 0.06
Nodes (28): SortedPathSet(), SortPaths(), sortablePath, newRemoveHasLocalModificationsError(), newRemoveHasStagedAndLocalStateError(), newRemoveHasStagedChangesError(), newRemovePathDidNotMatchError(), newRemoveRecursiveRequiredError() (+20 more)

### Community 1 - "Object Storage"
Cohesion: 0.07
Nodes (13): Compress(), Decompress(), NewObjectService(), ObjectService, Blob, NewBlob(), Object, DeserializeObject() (+5 more)

### Community 2 - "Index & FileMode"
Cohesion: 0.09
Nodes (9): FileMode, formatStoredFileMode(), NewFileMode(), NewFileModeFromOSMode(), NewFileModeFromTreeMode(), Index, NewLsFilesService(), LsFilesOptions (+1 more)

### Community 3 - "Path Types & Workspace"
Cohesion: 0.09
Nodes (15): AbsolutePath, NormalizedPath, isRelativeOutside(), NewAbsolutePath(), ParseNormalizedPath(), pathWithinDir(), validateNormalizedFormat(), Workspace (+7 more)

### Community 4 - "Commit & Identity"
Cohesion: 0.13
Nodes (19): Commit, cloneCommitFields(), DeserializeCommit(), deserializeFields(), deserializeFieldStr(), deserializeIdentity(), deserializeTreeOrParent(), isValidCommitField() (+11 more)

### Community 5 - "Commit Service & Hash"
Cohesion: 0.10
Nodes (14): NewCommitTreeService(), CommitTreeService, NewLogService(), LogEntry, LogOptions, LogService, Hash, NewHashFromBytes() (+6 more)

### Community 6 - "Tree Resolver"
Cohesion: 0.10
Nodes (9): PathHashes, NewTreeResolver(), NewTreeWalker(), TreeResolver, TreeWalker, WalkOptions, NewLsTreeService(), LsTreeOptions (+1 more)

### Community 7 - "Index Serialization"
Cohesion: 0.15
Nodes (12): NewIndexService(), IndexService, deserializeHeader(), DeserializeIndex(), deserializeIndexEntry(), NewEmptyIndex(), NewIndex(), NewIndexHeader() (+4 more)

### Community 8 - "Tree Operations"
Cohesion: 0.17
Nodes (17): NewTree(), NewTreeEntry(), NewTreeFromEntries(), parseTreeEntries(), serializeTreeEntries(), SortTreeEntries(), treeEntrySortKey(), validateTreeEntries() (+9 more)

### Community 9 - "Diff Engine"
Cohesion: 0.12
Nodes (13): ContentLoaderFunc, NewDiffService(), sortDiffResults(), DiffMode, DiffOptions, DiffResult, DiffService, DiffStatus (+5 more)

### Community 10 - "Config & Validation"
Cohesion: 0.14
Nodes (7): NewConfigService(), ConfigService, NewHashObjectService(), HashObjectOptions, HashObjectService, PathMustBeFile(), StringMustNotBeEmpty()

### Community 11 - "Documentation Hub"
Cohesion: 0.23
Nodes (16): Plumbing vs Porcelain Commands, ADD Command Guide, AI-Powered .NET Development — Comprehensive Learning Guide, CAT-FILE Command Guide, How to Write a Command Guide Document, CONFIG Command Guide, Documentation Guide, HASH-OBJECT Command Guide (+8 more)

### Community 12 - "Branch Switch"
Cohesion: 0.21
Nodes (10): pathState, changedPaths(), hasOverwriteConflict(), isSafeState(), isSamePathState(), newPathStateFromPathHashes(), NewSwitchService(), SwitchOptions (+2 more)

### Community 13 - "Path Resolver"
Cohesion: 0.25
Nodes (6): classifyPathspec(), NewPathResolver(), PathResolver, PathspecType, ResolvedPath, NewNormalizedPath()

### Community 14 - "Show Inspection"
Cohesion: 0.18
Nodes (10): resolvedObjectRef, NewShowService(), sortTreeEntries(), ShowBlobResult, ShowCommitResult, ShowMode, ShowOptions, ShowResult (+2 more)

### Community 15 - "Config Domain"
Cohesion: 0.23
Nodes (7): Config, cloneConfigSection(), NewConfig(), NewConfigFromSections(), validateConfigName(), ConfigEntry, ConfigSection

### Community 16 - "CLI Diff/Show"
Cohesion: 0.24
Nodes (9): printAddedFileHeader(), printDeletedFileHeader(), printDiffResults(), printHunks(), printModifiedFileHeader(), printShowBlob(), printShowCommit(), printShowResult() (+1 more)

### Community 17 - "Reset Service"
Cohesion: 0.23
Nodes (9): collectDeletePaths(), collectWritePaths(), NewResetService(), pruneEmptyParentDirs(), validateMode(), ResetMode, ResetOptions, ResetResult (+1 more)

### Community 18 - "Ref Service"
Cohesion: 0.30
Nodes (3): NewRefService(), validateRefPrefix(), RefService

### Community 19 - "Restore Service"
Cohesion: 0.26
Nodes (5): NewEmptyIndexEntry(), NewRestoreService(), RestoreMode, RestoreOptions, RestoreService

### Community 20 - "Branch Service"
Cohesion: 0.27
Nodes (4): NewBranchService(), validateBranchName(), BranchListItem, BranchService

### Community 21 - "Commit Resolver"
Cohesion: 0.33
Nodes (5): isFullHash(), NewCommitResolver(), parsePositiveInteger(), parseTildeExpression(), CommitResolver

### Community 22 - "Status Inspection"
Cohesion: 0.29
Nodes (7): FileStatus, collectStaged(), collectUnstaged(), collectUntracked(), NewStatusService(), StatusResult, StatusService

### Community 23 - "Myers Algorithm"
Cohesion: 0.28
Nodes (4): LineDiff, NewMyersDiffAlgorithm(), MyersDiffAlgorithm, OperationType

### Community 24 - "Design Rationales"
Cohesion: 0.25
Nodes (8): SHA-256 Hashing, Content-Addressed Object Storage, Index as Flat Staging List, Object Serialization Format (type size\0body), Three-Way Comparison Safety Checks for rm, Scoped Removal of Deleted Files in add, Three-Area Model (HEAD / Index / Working Tree), Canonical Tree Entry Sorting

### Community 25 - "Init Service"
Cohesion: 0.43
Nodes (4): directoryExists(), ensureDirectory(), ensureRegularFile(), InitService

### Community 26 - "Add Service"
Cohesion: 0.33
Nodes (4): NewAddService(), AddOptions, AddResult, AddService

### Community 27 - "CLI Root Wiring"
Cohesion: 0.29
Nodes (4): initializeServices(), NewRemoveService(), NewReadTreeService(), ReadTreeService

### Community 29 - "Update Index"
Cohesion: 0.43
Nodes (3): ComputeIndexFlags(), NewIndexEntry(), UpdateIndexService

### Community 30 - "CatFile Inspection"
Cohesion: 0.40
Nodes (3): NewCatFileService(), CatFileOptions, CatFileService

### Community 31 - "Change Detector"
Cohesion: 0.33
Nodes (4): NewChangeDetector(), ChangeDetector, ChangeResult, FileState

### Community 32 - "Error Handling"
Cohesion: 0.50
Nodes (5): Production-Ready Error Handling Guide for Go, Go CLI Error Handling Guide, CLI Layer as Error Rendering Boundary, Error Classification via Kind, Error Translation at Layer Boundaries

### Community 37 - "File Stat"
Cohesion: 0.67
Nodes (3): NewFileStatFromPath(), newFileStatFromSyscall(), FileStat

### Community 39 - "Defensive Patterns"
Cohesion: 0.67
Nodes (3): Defensive Copies for Immutability, Separate Path Domains (AbsolutePath vs NormalizedPath), Validation at Constructors and Parsers

## Knowledge Gaps
- **59 isolated node(s):** `ResetMode`, `ResetOptions`, `ResetResult`, `LogEntry`, `LogOptions` (+54 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `initializeServices()` connect `CLI Root Wiring` to `Object Storage`, `Index & FileMode`, `Path Types & Workspace`, `Commit Service & Hash`, `Tree Resolver`, `Index Serialization`, `Tree Operations`, `Diff Engine`, `Config & Validation`, `Branch Switch`, `Path Resolver`, `Show Inspection`, `Reset Service`, `Ref Service`, `Restore Service`, `Branch Service`, `Commit Resolver`, `Status Inspection`, `Myers Algorithm`, `Add Service`, `Update Ref`, `CatFile Inspection`, `Change Detector`, `Symbolic Ref`, `Config Storage`, `Index Storage`, `Commit Execution`, `Staging Update`?**
  _High betweenness centrality (0.371) - this node is a cross-community bridge._
- **Why does `NewHashFromHex()` connect `Commit Service & Hash` to `Object Storage`, `Commit & Identity`, `Index Serialization`, `Tree Operations`, `Show Inspection`, `Ref Service`, `Restore Service`, `Branch Service`, `Commit Resolver`?**
  _High betweenness centrality (0.166) - this node is a cross-community bridge._
- **Why does `deserializeIndexEntry()` connect `Index Serialization` to `Path Types & Workspace`, `Commit Service & Hash`?**
  _High betweenness centrality (0.124) - this node is a cross-community bridge._
- **Are the 34 inferred relationships involving `initializeServices()` (e.g. with `NewWorkspace()` and `NewObjectStorage()`) actually correct?**
  _`initializeServices()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **Are the 11 inferred relationships involving `NewHashFromHex()` (e.g. with `.resolveStartHash()` and `.CommitTree()`) actually correct?**
  _`NewHashFromHex()` has 11 INFERRED edges - model-reasoned connections that need verification._
- **What connects `ResetMode`, `ResetOptions`, `ResetResult` to the rest of the system?**
  _72 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Staging Removal` be split into smaller, more focused modules?**
  _Cohesion score 0.055904961565338925 - nodes in this community are weakly interconnected._