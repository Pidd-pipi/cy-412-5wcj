package constants

const (
	RepairStatusPending    = "pending"
	RepairStatusAssigned   = "assigned"
	RepairStatusProcessing = "processing"
	RepairStatusDone       = "done"
	RepairStatusAcceptance = "acceptance"
	RepairStatusClosed     = "closed"
)

var ValidRepairStatuses = map[string]bool{RepairStatusPending: true, RepairStatusAssigned: true, RepairStatusProcessing: true, RepairStatusDone: true, RepairStatusAcceptance: true, RepairStatusClosed: true}
