package entities

import "testing"

func TestPowerUpgradeChain(t *testing.T) {
	if got := Small.Upgrade(); got != Chef {
		t.Errorf("Small.Upgrade=%v want Chef", got)
	}
	if got := Chef.Upgrade(); got != Master {
		t.Errorf("Chef.Upgrade=%v want Master", got)
	}
	if got := Master.Upgrade(); got != Master {
		t.Errorf("Master.Upgrade should cap at Master, got %v", got)
	}
}

func TestPowerDowngradeChain(t *testing.T) {
	if np, died := Master.Downgrade(); np != Chef || died {
		t.Errorf("Master.Downgrade=(%v,%v) want (Chef,false)", np, died)
	}
	if np, died := Chef.Downgrade(); np != Small || died {
		t.Errorf("Chef.Downgrade=(%v,%v) want (Small,false)", np, died)
	}
	if np, died := Small.Downgrade(); np != Small || !died {
		t.Errorf("Small.Downgrade=(%v,%v) want (Small,true)", np, died)
	}
}

func TestPowerCapabilities(t *testing.T) {
	if Small.Big() || !Chef.Big() || !Master.Big() {
		t.Error("Big() wrong across tiers")
	}
	if Small.CanThrow() || Chef.CanThrow() || !Master.CanThrow() {
		t.Error("CanThrow() should only be true for Master")
	}
}
