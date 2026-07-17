// Package api is a thin adapter over the OpenAPI-generated apiclient:
// session factory (auth, timeout), HTTP error taxonomy (Classify), 429 retry
// transport, date parsing helper, connectivity probe, and entity facades
// (TaskController_* via ListTasks/GetTask/…; ProjectController_* via
// ListProjects/GetProject/CreateProject/UpdateProject/DeleteProject and
// ArchiveProject/TrashProject; TaskGroupController_* via
// ListSections/GetSection/CreateSection/UpdateSection/DeleteSection;
// ChecklistItemController_* via
// ListChecklistItems/GetChecklistItem/…; KanbanTaskStatusController_* via
// ListKanbanLinks/GetKanbanLink/CreateKanbanLink/UpdateKanbanLink/DeleteKanbanLink
// and MoveTaskToKanban).
package api
