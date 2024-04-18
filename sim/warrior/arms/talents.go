package arms

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/warrior"
)

func (war *ArmsWarrior) ApplyTalents() {
	specialAttacks := SpellMaskBladestorm | SpellMaskMortalStrike
	war.Warrior.ApplyCommonTalents(SpellMaskMortalStrike, SpellMaskMortalStrike, specialAttacks, SpellMaskMortalStrike)

	war.RegisterBladestorm()
	war.RegisterDeadlyCalm()
	war.RegisterSweepingStrikes()

	war.applyBloodFrenzy()
	war.applyImpale()
	war.applyImprovedSlam()
	war.applySlaughter()
	war.applySuddenDeath()
	war.applyTasteForBlood()
	war.applyWreckingCrew()

	// Apply glyphs after talents so we can modify spells added from talents
	war.ApplyGlyphs()
}

func (war *ArmsWarrior) applyTasteForBlood() {
	if war.Talents.TasteForBlood == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  warrior.SpellMaskOverpower,
		Kind:       core.SpellMod_BonusCrit_Rating,
		FloatValue: 20.0 * float64(war.Talents.TasteForBlood) * core.CritRatingPerCritChance,
	})

	procChance := []float64{0, 0.33, 0.66, 1}[war.Talents.TasteForBlood]

	// Use a specific aura for TfB so we can track procs
	// Overpower will check for any aura with the EnableOverpowerTag when it tries to cast
	tfbAura := war.RegisterAura(core.Aura{
		Label:    "Taste for Blood",
		ActionID: core.ActionID{SpellID: 60503},
		Duration: time.Second * 9,
		Tag:      warrior.EnableOverpowerTag,
	})

	core.MakePermanent(war.RegisterAura(core.Aura{
		Label: "Taste for Blood Trigger",
		Icd: &core.Cooldown{
			Timer:    war.NewTimer(),
			Duration: time.Second * 5,
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell != war.Rend {
				return
			}

			if !aura.Icd.IsReady(sim) {
				return
			}

			if sim.Proc(procChance, "Taste for Blood") {
				aura.Icd.Use(sim)
				tfbAura.Activate(sim)
			}
		},
	}))
}

func (war *ArmsWarrior) applyImpale() {
	if war.Talents.Impale == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskMortalStrike | warrior.SpellMaskSlam | warrior.SpellMaskOverpower,
		Kind:       core.SpellMod_BonusCrit_Rating,
		FloatValue: 0.1 * float64(war.Talents.Impale),
	})
}

func (war *ArmsWarrior) applyImprovedSlam() {
	if war.Talents.ImprovedSlam == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: warrior.SpellMaskSlam,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-500*war.Talents.ImprovedSlam),
	})

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  warrior.SpellMaskSlam,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.1 * float64(war.Talents.ImprovedSlam),
	})
}

func (war *ArmsWarrior) applySuddenDeath() {
	if war.Talents.SuddenDeath == 0 {
		return
	}

	procChance := 0.03 * float64(war.Talents.SuddenDeath)
	core.MakePermanent(war.RegisterAura(core.Aura{
		Label: "Sudden Death Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMelee) {
				return
			}

			if sim.Proc(procChance, "Sudden Death") {
				war.ColossusSmash.CD.Reset()
			}
		},
	}))
}

func (war *ArmsWarrior) TriggerSlaughter(sim *core.Simulation, target *core.Unit) {
	if war.Talents.LambsToTheSlaughter == 0 {
		return
	}

	rend := war.Rend.Dot(target)
	if rend != nil && rend.IsActive() {
		rend.Refresh(sim)
	}

	if !war.slaughter.IsActive() {
		war.slaughter.Activate(sim)
	} else {
		war.slaughter.Refresh(sim)
		war.slaughter.AddStack(sim)
	}
}

func (war *ArmsWarrior) applySlaughter() {
	if war.Talents.LambsToTheSlaughter == 0 {
		return
	}

	damageMod := war.AddDynamicMod(core.SpellModConfig{
		ClassMask:  SpellMaskMortalStrike | warrior.SpellMaskExecute | warrior.SpellMaskOverpower | warrior.SpellMaskSlam,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.0,
	})

	war.slaughter = war.RegisterAura(core.Aura{
		Label:     "Slaughter",
		ActionID:  core.ActionID{SpellID: 84586},
		Duration:  time.Second * 15,
		MaxStacks: war.Talents.LambsToTheSlaughter,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			bonus := 0.1 * float64(newStacks)
			damageMod.UpdateFloatValue(bonus)

			if newStacks != 0 {
				damageMod.Activate()
			} else {
				damageMod.Deactivate()
			}
		},
	})
}

func (war *ArmsWarrior) applyWreckingCrew() {
	if war.Talents.WreckingCrew == 0 {
		return
	}

	effect := 1.0 + (0.05 * float64(war.Talents.WreckingCrew))
	war.wreckingCrew = war.RegisterAura(core.Aura{
		Label:    "Wrecking Crew",
		ActionID: core.ActionID{SpellID: 56611},
		Duration: time.Second * 12,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			war.PseudoStats.SchoolDamageDealtMultiplier[core.SpellSchoolPhysical] *= effect
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			war.PseudoStats.SchoolDamageDealtMultiplier[core.SpellSchoolPhysical] /= effect
		},
	})

	core.RegisterPercentDamageModifierEffect(war.wreckingCrew, effect)
}

func (war *ArmsWarrior) TriggerWreckingCrew(sim *core.Simulation) {
	if war.Talents.WreckingCrew == 0 {
		return
	}

	procChance := 0.5 * float64(war.Talents.WreckingCrew)
	if sim.Proc(procChance, "Wrecking Crew") {
		war.wreckingCrew.Activate(sim)
	}
}
