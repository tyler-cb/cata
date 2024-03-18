package rogue

import (
	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (rogue *Rogue) registerStealthAura() {
	// TODO: Add Stealth spell for use with prepull in APL
	rogue.StealthAura = rogue.RegisterAura(core.Aura{
		Label:    "Stealth",
		ActionID: core.ActionID{SpellID: 1784},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			// Stealth triggered auras
			if rogue.Talents.Overkill {
				rogue.OverkillAura.Activate(sim)
			}
			if rogue.Spec == proto.Spec_SpecSubtletyRogue {
				rogue.MasterOfSubtletyAura.Activate(sim)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if rogue.Talents.Overkill {
				rogue.OverkillAura.Deactivate(sim)
				rogue.OverkillAura.Activate(sim)
			}
			if rogue.Spec == proto.Spec_SpecSubtletyRogue {
				rogue.MasterOfSubtletyAura.Deactivate(sim)
				rogue.MasterOfSubtletyAura.Activate(sim)
			}
		},
		// Stealth breaks on damage taken (if not absorbed)
		// This may be desirable later, but not applicable currently
	})
}
