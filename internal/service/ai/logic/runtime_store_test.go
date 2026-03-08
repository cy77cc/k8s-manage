package logic

import "testing"

func TestRuntimeStoreApprovalCRUD(t *testing.T) {
	t.Parallel()

	store := NewRuntimeStore(nil)
	store.CreateApprovalTask(&ApprovalTicket{ID: "apv-1", Tool: "service_restart", Status: "pending"})

	got, ok := store.GetApprovalTask("apv-1")
	if !ok || got.Tool != "service_restart" {
		t.Fatalf("GetApprovalTask() = %+v, %v", got, ok)
	}

	updated, ok := store.UpdateApprovalTask("apv-1", "approved")
	if !ok || updated.Status != "approved" {
		t.Fatalf("UpdateApprovalTask() = %+v, %v", updated, ok)
	}

	store.DeleteApprovalTask("apv-1")
	if _, ok := store.GetApprovalTask("apv-1"); ok {
		t.Fatal("DeleteApprovalTask() did not remove record")
	}
}
