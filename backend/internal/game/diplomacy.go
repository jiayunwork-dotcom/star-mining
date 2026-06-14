package game

import (
	"fmt"
	"math/rand"
	"sort"

	"star-mining/internal/models"
)

const (
	MaxAllianceMembers        = 3
	AllianceCooldownTurns     = 3
	TradeAgreementDuration    = 10
	MilitaryActionDeadline    = 2
	JointAttackBonus          = 1.2
	InitialDiplomacyValue     = 50.0
	MinDiplomacyValue         = 0.0
	MaxDiplomacyValue         = 100.0
	HostileThreshold          = 20.0
	FriendlyThreshold         = 70.0
	AllianceRelationCap       = 80.0
	WarDeclareThreshold       = 30.0
	SanctionRequiredSeconders = 2
	SanctionDuration          = 8
	SanctionProposalExpiry    = 2
	SanctionMaintenanceRate   = 0.05
	SanctionFeeMultiplier     = 2.0
)

var AllianceColors = []models.AllianceColor{
	models.AllianceColorRed,
	models.AllianceColorBlue,
	models.AllianceColorGreen,
	models.AllianceColorYellow,
	models.AllianceColorPurple,
	models.AllianceColorCyan,
}

func CreateAlliance(state *models.GameState, leaderID string, name string, color models.AllianceColor) (*models.Alliance, error) {
	if name == "" {
		return nil, fmt.Errorf("联盟名称不能为空")
	}

	existing := FindPlayerAlliance(state, leaderID)
	if existing != nil {
		return nil, fmt.Errorf("你已经在联盟中")
	}

	cd := FindPlayerCooldown(state, leaderID)
	if cd != nil {
		remaining := cd.CooldownTurns - (state.Turn - cd.LeftTurn)
		if remaining > 0 {
			return nil, fmt.Errorf("冷却期中，还需等待 %d 回合", remaining)
		}
	}

	for _, a := range state.Alliances {
		if a.Color == color && a.Status == models.AllianceActive {
			return nil, fmt.Errorf("该颜色已被其他联盟使用")
		}
	}

	validColor := false
	for _, c := range AllianceColors {
		if c == color {
			validColor = true
			break
		}
	}
	if !validColor {
		return nil, fmt.Errorf("无效的联盟颜色")
	}

	alliance := &models.Alliance{
		ID:           fmt.Sprintf("alliance-%d", len(state.Alliances)),
		Name:         name,
		LeaderID:     leaderID,
		MemberIDs:    []string{leaderID},
		Color:        color,
		Status:       models.AllianceActive,
		CreatedTurn:  state.Turn,
		InviteExpiry: make(map[string]int),
	}

	state.Alliances = append(state.Alliances, alliance)
	return alliance, nil
}

func SendAllianceInvite(state *models.GameState, allianceID string, inviterID string, targetID string) error {
	alliance := FindAllianceByID(state, allianceID)
	if alliance == nil {
		return fmt.Errorf("联盟不存在")
	}
	if alliance.LeaderID != inviterID {
		return fmt.Errorf("只有盟主可以邀请成员")
	}
	if alliance.Status != models.AllianceActive {
		return fmt.Errorf("联盟已解散")
	}
	if len(alliance.MemberIDs) >= MaxAllianceMembers {
		return fmt.Errorf("联盟已满员（最多%d人）", MaxAllianceMembers)
	}

	targetAlliance := FindPlayerAlliance(state, targetID)
	if targetAlliance != nil {
		return fmt.Errorf("该玩家已在其他联盟中")
	}

	cd := FindPlayerCooldown(state, targetID)
	if cd != nil {
		remaining := cd.CooldownTurns - (state.Turn - cd.LeftTurn)
		if remaining > 0 {
			return fmt.Errorf("该玩家处于冷却期")
		}
	}

	for _, m := range alliance.MemberIDs {
		if m == targetID {
			return fmt.Errorf("该玩家已在联盟中")
		}
	}

	alliance.InviteExpiry[targetID] = state.Turn
	return nil
}

func AcceptAllianceInvite(state *models.GameState, allianceID string, playerID string) error {
	alliance := FindAllianceByID(state, allianceID)
	if alliance == nil {
		return fmt.Errorf("联盟不存在")
	}

	if _, exists := alliance.InviteExpiry[playerID]; !exists {
		return fmt.Errorf("没有收到邀请")
	}

	if len(alliance.MemberIDs) >= MaxAllianceMembers {
		delete(alliance.InviteExpiry, playerID)
		return fmt.Errorf("联盟已满员")
	}

	targetAlliance := FindPlayerAlliance(state, playerID)
	if targetAlliance != nil {
		delete(alliance.InviteExpiry, playerID)
		return fmt.Errorf("你已在其他联盟中")
	}

	if IsPlayerSanctioned(state, playerID) {
		delete(alliance.InviteExpiry, playerID)
		return fmt.Errorf("被制裁的玩家不能加入联盟")
	}

	alliance.MemberIDs = append(alliance.MemberIDs, playerID)
	delete(alliance.InviteExpiry, playerID)
	return nil
}

func RejectAllianceInvite(state *models.GameState, allianceID string, playerID string) error {
	alliance := FindAllianceByID(state, allianceID)
	if alliance == nil {
		return fmt.Errorf("联盟不存在")
	}
	delete(alliance.InviteExpiry, playerID)
	return nil
}

func LeaveAlliance(state *models.GameState, playerID string) error {
	alliance := FindPlayerAlliance(state, playerID)
	if alliance == nil {
		return fmt.Errorf("你不在任何联盟中")
	}

	if alliance.LeaderID == playerID {
		return fmt.Errorf("盟主不能直接退出，请转让盟主或解散联盟")
	}

	alliance.MemberIDs = removeFromSlice(alliance.MemberIDs, playerID)

	state.PlayerCooldowns = append(state.PlayerCooldowns, &models.PlayerCooldown{
		PlayerID:      playerID,
		CooldownTurns: AllianceCooldownTurns,
		LeftTurn:      state.Turn,
	})

	CancelTradeAgreementsForPlayer(state, playerID)

	CancelMilitaryActionsForPlayer(state, playerID)

	if len(alliance.MemberIDs) <= 1 && alliance.LeaderID != playerID {
		alliance.Status = models.AllianceDisbanded
	}

	return nil
}

func KickAllianceMember(state *models.GameState, leaderID string, targetID string) error {
	alliance := FindPlayerAlliance(state, leaderID)
	if alliance == nil {
		return fmt.Errorf("你不在任何联盟中")
	}
	if alliance.LeaderID != leaderID {
		return fmt.Errorf("只有盟主可以踢人")
	}
	if targetID == leaderID {
		return fmt.Errorf("不能踢自己")
	}

	found := false
	for _, m := range alliance.MemberIDs {
		if m == targetID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("该玩家不在联盟中")
	}

	alliance.MemberIDs = removeFromSlice(alliance.MemberIDs, targetID)

	CancelTradeAgreementsForPlayer(state, targetID)
	CancelMilitaryActionsForPlayer(state, targetID)

	return nil
}

func DisbandAlliance(state *models.GameState, leaderID string) error {
	alliance := FindPlayerAlliance(state, leaderID)
	if alliance == nil {
		return fmt.Errorf("你不在任何联盟中")
	}
	if alliance.LeaderID != leaderID {
		return fmt.Errorf("只有盟主可以解散联盟")
	}

	for _, memberID := range alliance.MemberIDs {
		if memberID != leaderID {
			state.PlayerCooldowns = append(state.PlayerCooldowns, &models.PlayerCooldown{
				PlayerID:      memberID,
				CooldownTurns: AllianceCooldownTurns,
				LeftTurn:      state.Turn,
			})
		}
	}

	CancelAllTradeAgreementsForAlliance(state, alliance.ID)
	CancelAllMilitaryActionsForAlliance(state, alliance.ID)

	alliance.Status = models.AllianceDisbanded
	alliance.MemberIDs = nil
	alliance.InviteExpiry = make(map[string]int)
	return nil
}

func CreateTradeAgreement(state *models.GameState, playerID1 string, playerID2 string) (*models.TradeAgreement, error) {
	alliance1 := FindPlayerAlliance(state, playerID1)
	alliance2 := FindPlayerAlliance(state, playerID2)

	if alliance1 == nil || alliance2 == nil {
		return nil, fmt.Errorf("双方必须在同一联盟中才能签订贸易协定")
	}
	if alliance1.ID != alliance2.ID {
		return nil, fmt.Errorf("非联盟成员之间不能签订贸易协定")
	}

	if ArePlayersAtWar(state, playerID1, playerID2) {
		return nil, fmt.Errorf("交战期间不能签订贸易协定")
	}

	for _, ta := range state.TradeAgreements {
		if ta.Status == models.TradeAgreementActive {
			if (ta.PlayerID1 == playerID1 && ta.PlayerID2 == playerID2) ||
				(ta.PlayerID1 == playerID2 && ta.PlayerID2 == playerID1) {
				return nil, fmt.Errorf("双方已有生效的贸易协定")
			}
		}
	}

	agreement := &models.TradeAgreement{
		ID:          fmt.Sprintf("trade-%d-%s-%s", state.Turn, playerID1[:8], playerID2[:8]),
		PlayerID1:   playerID1,
		PlayerID2:   playerID2,
		AllianceID:  alliance1.ID,
		Status:      models.TradeAgreementActive,
		CreatedTurn: state.Turn,
		ExpiryTurn:  state.Turn + TradeAgreementDuration,
	}

	state.TradeAgreements = append(state.TradeAgreements, agreement)
	return agreement, nil
}

func RenewTradeAgreement(state *models.GameState, agreementID string) error {
	for _, ta := range state.TradeAgreements {
		if ta.ID == agreementID {
			if ta.Status != models.TradeAgreementActive {
				return fmt.Errorf("贸易协定已失效")
			}
			ta.ExpiryTurn = state.Turn + TradeAgreementDuration
			return nil
		}
	}
	return fmt.Errorf("贸易协定不存在")
}

func HasActiveTradeAgreement(state *models.GameState, playerID1 string, playerID2 string) bool {
	for _, ta := range state.TradeAgreements {
		if ta.Status == models.TradeAgreementActive {
			if (ta.PlayerID1 == playerID1 && ta.PlayerID2 == playerID2) ||
				(ta.PlayerID1 == playerID2 && ta.PlayerID2 == playerID1) {
				return true
			}
		}
	}
	return false
}

func GetTradeFeeMultiplier(state *models.GameState, buyerID string, sellerID string) float64 {
	mult := 1.0
	if HasActiveTradeAgreement(state, buyerID, sellerID) {
		mult = 0.5
	}
	if IsPlayerSanctioned(state, buyerID) {
		mult *= SanctionFeeMultiplier
	}
	return mult
}

func InitiateJointMilitaryAction(state *models.GameState, initiatorID string, targetPlayerID string, targetBodyID string) (*models.JointMilitaryAction, error) {
	alliance := FindPlayerAlliance(state, initiatorID)
	if alliance == nil {
		return nil, fmt.Errorf("你不在任何联盟中")
	}
	if alliance.LeaderID != initiatorID {
		return nil, fmt.Errorf("只有盟主可以发起联合军事行动")
	}

	if AreAllies(state, initiatorID, targetPlayerID) {
		return nil, fmt.Errorf("不能对盟友发起军事行动")
	}

	activeWar := FindActiveWarForAlliance(state, alliance.ID)
	if activeWar != nil {
		targetAlliance := FindPlayerAlliance(state, targetPlayerID)
		if targetAlliance == nil || (targetAlliance.ID != activeWar.AttackerAllianceID && targetAlliance.ID != activeWar.DefenderAllianceID) {
			return nil, fmt.Errorf("战争期间只能对交战联盟发起联合军事行动")
		}
	}

	participants := make([]*models.MilitaryParticipant, 0)
	for _, memberID := range alliance.MemberIDs {
		participants = append(participants, &models.MilitaryParticipant{
			PlayerID: memberID,
			FleetID:  "",
			Joined:   memberID == initiatorID,
		})
	}

	action := &models.JointMilitaryAction{
		ID:             fmt.Sprintf("military-%d-%s", state.Turn, alliance.ID),
		AllianceID:     alliance.ID,
		InitiatorID:    initiatorID,
		TargetPlayerID: targetPlayerID,
		TargetBodyID:   targetBodyID,
		Status:         models.MilitaryActionRecruiting,
		Participants:   participants,
		CreatedTurn:    state.Turn,
		DeadlineTurn:   state.Turn + MilitaryActionDeadline,
		ArrivalTurn:    0,
		TotalAttack:    0,
	}

	state.JointMilitaryActions = append(state.JointMilitaryActions, action)
	return action, nil
}

func JoinMilitaryAction(state *models.JointMilitaryAction, playerID string, fleetID string) error {
	if state.Status != models.MilitaryActionRecruiting {
		return fmt.Errorf("军事行动不在招募阶段")
	}

	for _, p := range state.Participants {
		if p.PlayerID == playerID {
			p.FleetID = fleetID
			p.Joined = true
			return nil
		}
	}
	return fmt.Errorf("你不是该行动的参与者")
}

func DeclineMilitaryAction(action *models.JointMilitaryAction, playerID string) error {
	if action.Status != models.MilitaryActionRecruiting {
		return fmt.Errorf("军事行动不在招募阶段")
	}

	for _, p := range action.Participants {
		if p.PlayerID == playerID {
			p.Joined = false
			p.FleetID = ""
			return nil
		}
	}
	return fmt.Errorf("你不是该行动的参与者")
}

func InitDiplomacyRelations(state *models.GameState) {
	existing := make(map[string]bool)
	for _, r := range state.DiplomacyRelations {
		key := r.Player1ID + ":" + r.Player2ID
		existing[key] = true
	}

	for i, p1 := range state.Players {
		for j, p2 := range state.Players {
			if i >= j {
				continue
			}
			key1 := p1.ID + ":" + p2.ID
			key2 := p2.ID + ":" + p1.ID
			if existing[key1] || existing[key2] {
				continue
			}

			state.DiplomacyRelations = append(state.DiplomacyRelations, &models.DiplomacyRelation{
				Player1ID: p1.ID,
				Player2ID: p2.ID,
				Value:     InitialDiplomacyValue,
			})
		}
	}
}

func GetDiplomacyValue(state *models.GameState, player1ID string, player2ID string) float64 {
	for _, r := range state.DiplomacyRelations {
		if (r.Player1ID == player1ID && r.Player2ID == player2ID) ||
			(r.Player1ID == player2ID && r.Player2ID == player1ID) {
			return r.Value
		}
	}
	return InitialDiplomacyValue
}

func ModifyDiplomacyValue(state *models.GameState, player1ID string, player2ID string, delta float64, reason string) *models.DiplomacyChange {
	if ArePlayersAtWar(state, player1ID, player2ID) {
		return nil
	}

	for _, r := range state.DiplomacyRelations {
		if (r.Player1ID == player1ID && r.Player2ID == player2ID) ||
			(r.Player1ID == player2ID && r.Player2ID == player1ID) {
			oldVal := r.Value
			r.Value += delta
			if r.Value < MinDiplomacyValue {
				r.Value = MinDiplomacyValue
			}
			if r.Value > MaxDiplomacyValue {
				r.Value = MaxDiplomacyValue
			}
			return &models.DiplomacyChange{
				PlayerID: player2ID,
				OldValue: oldVal,
				NewValue: r.Value,
				Change:   r.Value - oldVal,
				Reason:   reason,
			}
		}
	}
	return nil
}

func GetDiplomacyStatus(value float64) string {
	if value < HostileThreshold {
		return "hostile"
	}
	if value > FriendlyThreshold {
		return "friendly"
	}
	return "neutral"
}

func AreAllies(state *models.GameState, player1ID string, player2ID string) bool {
	a1 := FindPlayerAlliance(state, player1ID)
	if a1 == nil {
		return false
	}
	a2 := FindPlayerAlliance(state, player2ID)
	if a2 == nil {
		return false
	}
	return a1.ID == a2.ID
}

func FindAllianceByID(state *models.GameState, allianceID string) *models.Alliance {
	for _, a := range state.Alliances {
		if a.ID == allianceID {
			return a
		}
	}
	return nil
}

func FindPlayerAlliance(state *models.GameState, playerID string) *models.Alliance {
	for _, a := range state.Alliances {
		if a.Status != models.AllianceActive {
			continue
		}
		for _, m := range a.MemberIDs {
			if m == playerID {
				return a
			}
		}
	}
	return nil
}

func FindPlayerCooldown(state *models.GameState, playerID string) *models.PlayerCooldown {
	for _, cd := range state.PlayerCooldowns {
		if cd.PlayerID == playerID {
			remaining := cd.CooldownTurns - (state.Turn - cd.LeftTurn)
			if remaining > 0 {
				return cd
			}
		}
	}
	return nil
}

func CancelTradeAgreementsForPlayer(state *models.GameState, playerID string) {
	for _, ta := range state.TradeAgreements {
		if ta.Status == models.TradeAgreementActive {
			if ta.PlayerID1 == playerID || ta.PlayerID2 == playerID {
				ta.Status = models.TradeAgreementExpired
			}
		}
	}
}

func CancelAllTradeAgreementsForAlliance(state *models.GameState, allianceID string) {
	for _, ta := range state.TradeAgreements {
		if ta.AllianceID == allianceID && ta.Status == models.TradeAgreementActive {
			ta.Status = models.TradeAgreementExpired
		}
	}
}

func CancelMilitaryActionsForPlayer(state *models.GameState, playerID string) {
	for _, action := range state.JointMilitaryActions {
		if action.Status != models.MilitaryActionRecruiting && action.Status != models.MilitaryActionInProgress {
			continue
		}
		if action.InitiatorID == playerID {
			action.Status = models.MilitaryActionCancelled
		}
	}
}

func CancelAllMilitaryActionsForAlliance(state *models.GameState, allianceID string) {
	for _, action := range state.JointMilitaryActions {
		if action.AllianceID == allianceID {
			if action.Status == models.MilitaryActionRecruiting || action.Status == models.MilitaryActionInProgress {
				action.Status = models.MilitaryActionCancelled
			}
		}
	}
}

func ProcessDiplomacyTurn(state *models.GameState) ([]*models.DiplomacyChange, []*models.SanctionEvent) {
	changes := make([]*models.DiplomacyChange, 0)
	var sanctionEvents []*models.SanctionEvent

	for _, alliance := range state.Alliances {
		if alliance.Status != models.AllianceActive {
			continue
		}
		for i := 0; i < len(alliance.MemberIDs); i++ {
			for j := i + 1; j < len(alliance.MemberIDs); j++ {
				if ArePlayersAtWar(state, alliance.MemberIDs[i], alliance.MemberIDs[j]) {
					continue
				}
				currentValue := GetDiplomacyValue(state, alliance.MemberIDs[i], alliance.MemberIDs[j])
				if currentValue >= AllianceRelationCap {
					continue
				}
				change := ModifyDiplomacyValue(state, alliance.MemberIDs[i], alliance.MemberIDs[j], 1.0, "联盟每回合自动增加")
				if change != nil && change.Change > 0 {
					for _, r := range state.DiplomacyRelations {
						if (r.Player1ID == alliance.MemberIDs[i] && r.Player2ID == alliance.MemberIDs[j]) ||
							(r.Player1ID == alliance.MemberIDs[j] && r.Player2ID == alliance.MemberIDs[i]) {
							if r.Value > AllianceRelationCap {
								r.Value = AllianceRelationCap
								change.NewValue = r.Value
								change.Change = r.Value - change.OldValue
							}
							break
						}
					}
					if change.Change > 0 {
						changes = append(changes, change)
					}
				}
			}
		}
	}

	for _, war := range state.AllianceWars {
		if war.Status != models.WarActive {
			continue
		}
		cancelTradeAgreementsBetweenAlliances(state, war.AttackerAllianceID, war.DefenderAllianceID)
	}

	for _, ta := range state.TradeAgreements {
		if ta.Status == models.TradeAgreementActive && state.Turn >= ta.ExpiryTurn {
			ta.Status = models.TradeAgreementExpired
		}
	}

	for _, action := range state.JointMilitaryActions {
		if action.Status == models.MilitaryActionRecruiting && state.Turn >= action.DeadlineTurn {
			joinedCount := 0
			for _, p := range action.Participants {
				if p.Joined && p.FleetID != "" {
					joinedCount++
				}
			}
			if joinedCount > 0 {
				action.Status = models.MilitaryActionInProgress
				action.ArrivalTurn = state.Turn + 2
			} else {
				action.Status = models.MilitaryActionCancelled
			}
		}
	}

	ProcessExpiredSanctionProposals(state)
	sanctionEvents = ProcessActiveSanctions(state)

	ProcessExpiredInvites(state)

	HandleBankruptMembers(state)

	CleanCooldowns(state)

	return changes, sanctionEvents
}

func ProcessJointMilitaryActions(state *models.GameState, playerMap map[string]*models.Player, rng *rand.Rand) []*models.CombatRecord {
	records := make([]*models.CombatRecord, 0)

	for _, action := range state.JointMilitaryActions {
		if action.Status != models.MilitaryActionInProgress {
			continue
		}
		if state.Turn < action.ArrivalTurn {
			continue
		}

		totalAttack := 0.0
		participatingFleets := make([]*models.Fleet, 0)

		for _, p := range action.Participants {
			if !p.Joined || p.FleetID == "" {
				continue
			}
			player := playerMap[p.PlayerID]
			if player == nil {
				continue
			}
			for _, fleet := range player.Fleets {
				if fleet.ID == p.FleetID {
					techBonus := GetCombatBonus(player.TechTree)
					totalAttack += fleet.TotalAttack * techBonus
					participatingFleets = append(participatingFleets, fleet)
					break
				}
			}
		}

		if totalAttack <= 0 {
			action.Status = models.MilitaryActionCancelled
			continue
		}

		totalAttack *= JointAttackBonus
		action.TotalAttack = totalAttack

		defender := playerMap[action.TargetPlayerID]
		if defender == nil || defender.IsDefeated || defender.IsBankrupt {
			action.Status = models.MilitaryActionCompleted
			continue
		}

		defenderFleet := &models.Fleet{
			ID:            "defender-combined",
			Name:          "防御联合舰队",
			PlayerID:      action.TargetPlayerID,
			Ships:         make([]*models.Ship, 0),
			TotalCargo:    make(models.Resources),
		}

		if action.TargetBodyID != "" {
			for _, fleet := range defender.Fleets {
				if fleet.CurrentBodyID == action.TargetBodyID && !fleet.IsMoving && len(fleet.Ships) > 0 {
					defenderFleet.Ships = append(defenderFleet.Ships, fleet.Ships...)
				}
			}

			defenderAlliance := FindPlayerAlliance(state, action.TargetPlayerID)
			if defenderAlliance != nil {
				for _, allyID := range defenderAlliance.MemberIDs {
					if allyID == action.TargetPlayerID {
						continue
					}
					ally := playerMap[allyID]
					if ally == nil || ally.IsDefeated || ally.IsBankrupt {
						continue
					}
					allyAlliance := FindPlayerAlliance(state, allyID)
					if allyAlliance == nil || allyAlliance.ID != defenderAlliance.ID {
						continue
					}
					for _, fleet := range ally.Fleets {
						if fleet.CurrentBodyID == "" || fleet.CurrentBodyID != action.TargetBodyID {
							continue
						}
						if fleet.IsMoving {
							continue
						}
						if len(fleet.Ships) == 0 {
							continue
						}
						defenderFleet.Ships = append(defenderFleet.Ships, fleet.Ships...)
					}
				}
			}
		}

		UpdateFleetStats(defenderFleet)

		attackerFleet := &models.Fleet{
			ID:         "attacker-combined",
			Name:       "联合攻击舰队",
			PlayerID:   action.InitiatorID,
			Ships:      make([]*models.Ship, 0),
			TotalCargo: make(models.Resources),
		}
		for _, f := range participatingFleets {
			attackerFleet.Ships = append(attackerFleet.Ships, f.Ships...)
		}
		UpdateFleetStats(attackerFleet)

		defenderTechBonus := GetCombatBonus(defender.TechTree)
		combatResult := SimulateCombat(attackerFleet, defenderFleet, 1.0, defenderTechBonus, rng)

		record := &models.CombatRecord{
			CombatID:        action.ID,
			AttackerID:      action.InitiatorID,
			AttackerName:    "联合舰队",
			DefenderID:      action.TargetPlayerID,
			DefenderName:    defender.Name,
			Winner:          combatResult.Winner,
			AttackerLosses:  len(combatResult.AttackerLosses),
			DefenderLosses:  len(combatResult.DefenderLosses),
			AttackerDamage:  combatResult.AttackerDamage,
			DefenderDamage:  combatResult.DefenderDamage,
		}

		records = append(records, record)

		ModifyDiplomacyValue(state, action.InitiatorID, action.TargetPlayerID, -30.0, "联合军事攻击")

		action.Status = models.MilitaryActionCompleted

		survivingShipIDs := make(map[string]bool)
		for _, s := range attackerFleet.Ships {
			survivingShipIDs[s.ID] = true
		}
		for _, f := range participatingFleets {
			surviving := make([]*models.Ship, 0)
			for _, s := range f.Ships {
				if survivingShipIDs[s.ID] {
					surviving = append(surviving, s)
				}
			}
			f.Ships = surviving
			UpdateFleetStats(f)
		}
	}

	return records
}

func ProcessWarAutoCombat(state *models.GameState, playerMap map[string]*models.Player, rng *rand.Rand) []*models.CombatRecord {
	records := make([]*models.CombatRecord, 0)

	for _, war := range state.AllianceWars {
		if war.Status != models.WarActive {
			continue
		}

		attackerFleets := make([]*models.Fleet, 0)
		defenderFleets := make([]*models.Fleet, 0)

		attackerAlliance := FindAllianceByID(state, war.AttackerAllianceID)
		defenderAlliance := FindAllianceByID(state, war.DefenderAllianceID)
		if attackerAlliance == nil || defenderAlliance == nil {
			continue
		}

		attackerMemberSet := make(map[string]bool)
		for _, mid := range attackerAlliance.MemberIDs {
			attackerMemberSet[mid] = true
		}
		defenderMemberSet := make(map[string]bool)
		for _, mid := range defenderAlliance.MemberIDs {
			defenderMemberSet[mid] = true
		}

		for _, player := range state.Players {
			if player.IsDefeated || player.IsBankrupt {
				continue
			}
			isAttacker := attackerMemberSet[player.ID]
			isDefender := defenderMemberSet[player.ID]
			if !isAttacker && !isDefender {
				continue
			}
			for _, fleet := range player.Fleets {
				if fleet.IsMoving || len(fleet.Ships) == 0 {
					continue
				}
				if fleet.CurrentBodyID == "" {
					continue
				}
				if isAttacker {
					attackerFleets = append(attackerFleets, fleet)
				} else {
					defenderFleets = append(defenderFleets, fleet)
				}
			}
		}

		attackerByBody := make(map[string][]*models.Fleet)
		defenderByBody := make(map[string][]*models.Fleet)

		for _, f := range attackerFleets {
			attackerByBody[f.CurrentBodyID] = append(attackerByBody[f.CurrentBodyID], f)
		}
		for _, f := range defenderFleets {
			defenderByBody[f.CurrentBodyID] = append(defenderByBody[f.CurrentBodyID], f)
		}

		for bodyID, aFleets := range attackerByBody {
			dFleets, exists := defenderByBody[bodyID]
			if !exists || len(dFleets) == 0 {
				continue
			}

			attackerCombined := &models.Fleet{
				ID:         "war-attacker-" + bodyID,
				Name:       "攻击方联合舰队",
				PlayerID:   war.AttackerAllianceID,
				Ships:      make([]*models.Ship, 0),
				TotalCargo: make(models.Resources),
			}
			for _, f := range aFleets {
				attackerCombined.Ships = append(attackerCombined.Ships, f.Ships...)
			}
			UpdateFleetStats(attackerCombined)

			defenderCombined := &models.Fleet{
				ID:         "war-defender-" + bodyID,
				Name:       "防御方联合舰队",
				PlayerID:   war.DefenderAllianceID,
				Ships:      make([]*models.Ship, 0),
				TotalCargo: make(models.Resources),
			}
			for _, f := range dFleets {
				defenderCombined.Ships = append(defenderCombined.Ships, f.Ships...)
			}
			UpdateFleetStats(defenderCombined)

			attackerTechBonus := 1.0
			for _, f := range aFleets {
				p := playerMap[f.PlayerID]
				if p != nil {
					bonus := GetCombatBonus(p.TechTree)
					if bonus > attackerTechBonus {
						attackerTechBonus = bonus
					}
				}
			}

			defenderTechBonus := 1.0
			for _, f := range dFleets {
				p := playerMap[f.PlayerID]
				if p != nil {
					bonus := GetCombatBonus(p.TechTree)
					if bonus > defenderTechBonus {
						defenderTechBonus = bonus
					}
				}
			}

			combatResult := SimulateCombat(attackerCombined, defenderCombined, attackerTechBonus, defenderTechBonus, rng)

			record := &models.CombatRecord{
				CombatID:       fmt.Sprintf("war-combat-%d-%s", state.Turn, bodyID),
				AttackerID:     war.AttackerAllianceID,
				AttackerName:   attackerAlliance.Name,
				DefenderID:     war.DefenderAllianceID,
				DefenderName:   defenderAlliance.Name,
				Winner:         combatResult.Winner,
				AttackerLosses: len(combatResult.AttackerLosses),
				DefenderLosses: len(combatResult.DefenderLosses),
				AttackerDamage: combatResult.AttackerDamage,
				DefenderDamage: combatResult.DefenderDamage,
			}
			records = append(records, record)

			survivingAttackerShips := make(map[string]bool)
			for _, s := range attackerCombined.Ships {
				survivingAttackerShips[s.ID] = true
			}
			for _, f := range aFleets {
				surviving := make([]*models.Ship, 0)
				for _, s := range f.Ships {
					if survivingAttackerShips[s.ID] {
						surviving = append(surviving, s)
					}
				}
				f.Ships = surviving
				UpdateFleetStats(f)
			}

			survivingDefenderShips := make(map[string]bool)
			for _, s := range defenderCombined.Ships {
				survivingDefenderShips[s.ID] = true
			}
			for _, f := range dFleets {
				surviving := make([]*models.Ship, 0)
				for _, s := range f.Ships {
					if survivingDefenderShips[s.ID] {
						surviving = append(surviving, s)
					}
				}
				f.Ships = surviving
				UpdateFleetStats(f)
			}
		}
	}

	return records
}

func ProcessExpiredInvites(state *models.GameState) {
	for _, alliance := range state.Alliances {
		if alliance.Status != models.AllianceActive {
			continue
		}
		for playerID, inviteTurn := range alliance.InviteExpiry {
			if state.Turn > inviteTurn {
				delete(alliance.InviteExpiry, playerID)
			}
		}
	}
}

func HandleBankruptMembers(state *models.GameState) {
	for _, alliance := range state.Alliances {
		if alliance.Status != models.AllianceActive {
			continue
		}

		removed := make([]string, 0)
		activeMembers := make([]string, 0)

		for _, memberID := range alliance.MemberIDs {
			member := findPlayerInState(state, memberID)
			if member == nil || member.IsBankrupt || member.IsDefeated {
				removed = append(removed, memberID)
			} else {
				activeMembers = append(activeMembers, memberID)
			}
		}

		if len(removed) > 0 {
			alliance.MemberIDs = activeMembers

			for _, rmid := range removed {
				CancelTradeAgreementsForPlayer(state, rmid)
			}

			if alliance.LeaderID != "" {
				leaderStillMember := false
				for _, m := range alliance.MemberIDs {
					if m == alliance.LeaderID {
						leaderStillMember = true
						break
					}
				}

				if !leaderStillMember {
					if len(alliance.MemberIDs) > 0 {
						bestScore := -1.0
						newLeader := alliance.MemberIDs[0]
						for _, m := range alliance.MemberIDs {
							p := findPlayerInState(state, m)
							if p != nil {
								score := CalculateScore(p, state.Exchanges)
								if score > bestScore {
									bestScore = score
									newLeader = m
								}
							}
						}
						alliance.LeaderID = newLeader
					}
				}
			}

			if len(alliance.MemberIDs) <= 1 {
				alliance.Status = models.AllianceDisbanded
				CancelAllTradeAgreementsForAlliance(state, alliance.ID)
				CancelAllMilitaryActionsForAlliance(state, alliance.ID)
			}
		}
	}
}

func CleanCooldowns(state *models.GameState) {
	active := make([]*models.PlayerCooldown, 0)
	for _, cd := range state.PlayerCooldowns {
		remaining := cd.CooldownTurns - (state.Turn - cd.LeftTurn)
		if remaining > 0 {
			active = append(active, cd)
		}
	}
	state.PlayerCooldowns = active
}

func findPlayerInState(state *models.GameState, playerID string) *models.Player {
	for _, p := range state.Players {
		if p.ID == playerID {
			return p
		}
	}
	return nil
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func GetAvailableAllianceColors(state *models.GameState) []models.AllianceColor {
	usedColors := make(map[models.AllianceColor]bool)
	for _, a := range state.Alliances {
		if a.Status == models.AllianceActive {
			usedColors[a.Color] = true
		}
	}
	available := make([]models.AllianceColor, 0)
	for _, c := range AllianceColors {
		if !usedColors[c] {
			available = append(available, c)
		}
	}
	return available
}

func GetPlayerRelations(state *models.GameState, playerID string) []*models.DiplomacyRelation {
	relations := make([]*models.DiplomacyRelation, 0)
	for _, r := range state.DiplomacyRelations {
		if r.Player1ID == playerID || r.Player2ID == playerID {
			relations = append(relations, &models.DiplomacyRelation{
				Player1ID: r.Player1ID,
				Player2ID: r.Player2ID,
				Value:     r.Value,
			})
		}
	}
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Value > relations[j].Value
	})
	return relations
}

func TransferLeadership(state *models.GameState, currentLeaderID string, newLeaderID string) error {
	alliance := FindPlayerAlliance(state, currentLeaderID)
	if alliance == nil {
		return fmt.Errorf("你不在任何联盟中")
	}
	if alliance.LeaderID != currentLeaderID {
		return fmt.Errorf("只有盟主可以转让领导权")
	}

	found := false
	for _, m := range alliance.MemberIDs {
		if m == newLeaderID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("目标玩家不在联盟中")
	}

	alliance.LeaderID = newLeaderID
	return nil
}

func DeclareWar(state *models.GameState, initiatorID string, targetAllianceID string) (*models.AllianceWar, error) {
	initiatorAlliance := FindPlayerAlliance(state, initiatorID)
	if initiatorAlliance == nil {
		return nil, fmt.Errorf("你不在任何联盟中")
	}
	if initiatorAlliance.LeaderID != initiatorID {
		return nil, fmt.Errorf("只有盟主可以宣战")
	}

	targetAlliance := FindAllianceByID(state, targetAllianceID)
	if targetAlliance == nil || targetAlliance.Status != models.AllianceActive {
		return nil, fmt.Errorf("目标联盟不存在或已解散")
	}

	if initiatorAlliance.ID == targetAllianceID {
		return nil, fmt.Errorf("不能对自己联盟宣战")
	}

	activeWar := FindActiveWarBetweenAlliances(state, initiatorAlliance.ID, targetAllianceID)
	if activeWar != nil {
		return nil, fmt.Errorf("双方已在战争中")
	}

	avgDiplomacy := calculateAverageDiplomacyBetweenAlliances(state, initiatorAlliance, targetAlliance)
	if avgDiplomacy >= WarDeclareThreshold {
		return nil, fmt.Errorf("外交关系平均值(%.1f)未低于%.0f，无法宣战", avgDiplomacy, WarDeclareThreshold)
	}

	war := &models.AllianceWar{
		ID:                  fmt.Sprintf("war-%d-%s-vs-%s", state.Turn, initiatorAlliance.ID, targetAllianceID),
		AttackerAllianceID:  initiatorAlliance.ID,
		DefenderAllianceID:  targetAllianceID,
		DeclaredTurn:        state.Turn,
		Status:              models.WarActive,
		AttackerTotalAssets: calculateAllianceTotalAssets(state, initiatorAlliance),
		DefenderTotalAssets: calculateAllianceTotalAssets(state, targetAlliance),
	}

	state.AllianceWars = append(state.AllianceWars, war)

	lockDiplomacyBetweenAlliances(state, initiatorAlliance, targetAlliance)

	return war, nil
}

func SurrenderWar(state *models.GameState, surrenderLeaderID string, warID string) ([]*models.ReparationDetail, error) {
	surrenderAlliance := FindPlayerAlliance(state, surrenderLeaderID)
	if surrenderAlliance == nil {
		return nil, fmt.Errorf("你不在任何联盟中")
	}
	if surrenderAlliance.LeaderID != surrenderLeaderID {
		return nil, fmt.Errorf("只有盟主可以投降")
	}

	var war *models.AllianceWar
	for _, w := range state.AllianceWars {
		if w.ID == warID && w.Status == models.WarActive {
			war = w
			break
		}
	}
	if war == nil {
		return nil, fmt.Errorf("战争不存在或已结束")
	}

	if war.AttackerAllianceID != surrenderAlliance.ID && war.DefenderAllianceID != surrenderAlliance.ID {
		return nil, fmt.Errorf("你的联盟不是交战方")
	}

	var winnerAllianceID string
	if war.AttackerAllianceID == surrenderAlliance.ID {
		winnerAllianceID = war.DefenderAllianceID
	} else {
		winnerAllianceID = war.AttackerAllianceID
	}

	winnerAlliance := FindAllianceByID(state, winnerAllianceID)
	if winnerAlliance == nil {
		return nil, fmt.Errorf("胜方联盟不存在")
	}

	reparations := calculateReparations(state, surrenderAlliance, winnerAlliance)

	war.Status = models.WarSurrendered
	war.SurrenderedAllianceID = surrenderAlliance.ID
	war.SurrenderTurn = state.Turn

	return reparations, nil
}

func ArePlayersAtWar(state *models.GameState, player1ID string, player2ID string) bool {
	a1 := FindPlayerAlliance(state, player1ID)
	if a1 == nil {
		return false
	}
	a2 := FindPlayerAlliance(state, player2ID)
	if a2 == nil {
		return false
	}
	if a1.ID == a2.ID {
		return false
	}
	return FindActiveWarBetweenAlliances(state, a1.ID, a2.ID) != nil
}

func FindActiveWarBetweenAlliances(state *models.GameState, allianceID1 string, allianceID2 string) *models.AllianceWar {
	for _, war := range state.AllianceWars {
		if war.Status != models.WarActive {
			continue
		}
		if (war.AttackerAllianceID == allianceID1 && war.DefenderAllianceID == allianceID2) ||
			(war.AttackerAllianceID == allianceID2 && war.DefenderAllianceID == allianceID1) {
			return war
		}
	}
	return nil
}

func FindActiveWarForAlliance(state *models.GameState, allianceID string) *models.AllianceWar {
	for _, war := range state.AllianceWars {
		if war.Status != models.WarActive {
			continue
		}
		if war.AttackerAllianceID == allianceID || war.DefenderAllianceID == allianceID {
			return war
		}
	}
	return nil
}

func FindActiveWarForPlayer(state *models.GameState, playerID string) *models.AllianceWar {
	alliance := FindPlayerAlliance(state, playerID)
	if alliance == nil {
		return nil
	}
	return FindActiveWarForAlliance(state, alliance.ID)
}

func calculateAverageDiplomacyBetweenAlliances(state *models.GameState, alliance1 *models.Alliance, alliance2 *models.Alliance) float64 {
	totalValue := 0.0
	pairCount := 0

	for _, m1 := range alliance1.MemberIDs {
		for _, m2 := range alliance2.MemberIDs {
			totalValue += GetDiplomacyValue(state, m1, m2)
			pairCount++
		}
	}

	if pairCount == 0 {
		return InitialDiplomacyValue
	}
	return totalValue / float64(pairCount)
}

func lockDiplomacyBetweenAlliances(state *models.GameState, alliance1 *models.Alliance, alliance2 *models.Alliance) {
	for _, m1 := range alliance1.MemberIDs {
		for _, m2 := range alliance2.MemberIDs {
			for _, r := range state.DiplomacyRelations {
				if (r.Player1ID == m1 && r.Player2ID == m2) ||
					(r.Player1ID == m2 && r.Player2ID == m1) {
					r.Value = 0
					break
				}
			}
		}
	}
}

func cancelTradeAgreementsBetweenAlliances(state *models.GameState, allianceID1 string, allianceID2 string) {
	a1 := FindAllianceByID(state, allianceID1)
	a2 := FindAllianceByID(state, allianceID2)
	if a1 == nil || a2 == nil {
		return
	}

	memberSet := make(map[string]bool)
	for _, m := range a2.MemberIDs {
		memberSet[m] = true
	}

	for _, ta := range state.TradeAgreements {
		if ta.Status != models.TradeAgreementActive {
			continue
		}
		if ta.AllianceID != allianceID1 && ta.AllianceID != allianceID2 {
			p1InA1 := false
			p2InA1 := false
			p1InA2 := memberSet[ta.PlayerID1]
			p2InA2 := memberSet[ta.PlayerID2]
			for _, m := range a1.MemberIDs {
				if m == ta.PlayerID1 {
					p1InA1 = true
				}
				if m == ta.PlayerID2 {
					p2InA1 = true
				}
			}
			if (p1InA1 && p2InA2) || (p2InA1 && p1InA2) {
				ta.Status = models.TradeAgreementExpired
			}
			continue
		}
		if ta.AllianceID == allianceID1 {
			if memberSet[ta.PlayerID1] || memberSet[ta.PlayerID2] {
				ta.Status = models.TradeAgreementExpired
			}
		}
		if ta.AllianceID == allianceID2 {
			isInA1 := false
			for _, m := range a1.MemberIDs {
				if m == ta.PlayerID1 || m == ta.PlayerID2 {
					isInA1 = true
					break
				}
			}
			if isInA1 {
				ta.Status = models.TradeAgreementExpired
			}
		}
	}
}

func calculateReparations(state *models.GameState, loserAlliance *models.Alliance, winnerAlliance *models.Alliance) []*models.ReparationDetail {
	reparations := make([]*models.ReparationDetail, 0)

	individualPayments := make(map[string]float64)
	for _, memberID := range loserAlliance.MemberIDs {
		player := findPlayerInState(state, memberID)
		if player == nil || player.IsBankrupt || player.IsDefeated {
			continue
		}
		payment := player.Credits * 0.10
		individualPayments[memberID] = payment
		player.Credits -= payment
	}

	totalPool := 0.0
	for _, amt := range individualPayments {
		totalPool += amt
	}

	activeWinners := make([]string, 0)
	for _, memberID := range winnerAlliance.MemberIDs {
		player := findPlayerInState(state, memberID)
		if player == nil || player.IsBankrupt || player.IsDefeated {
			continue
		}
		activeWinners = append(activeWinners, memberID)
	}

	if len(activeWinners) == 0 || totalPool == 0 {
		return reparations
	}

	sharePerWinner := totalPool / float64(len(activeWinners))
	for _, winnerID := range activeWinners {
		player := findPlayerInState(state, winnerID)
		if player != nil {
			player.Credits += sharePerWinner
		}
	}

	for loserID, payment := range individualPayments {
		loserPlayer := findPlayerInState(state, loserID)
		loserName := loserID
		if loserPlayer != nil {
			loserName = loserPlayer.Name
		}
		paymentPerWinner := payment / float64(len(activeWinners))
		for _, winnerID := range activeWinners {
			winnerPlayer := findPlayerInState(state, winnerID)
			winnerName := winnerID
			if winnerPlayer != nil {
				winnerName = winnerPlayer.Name
			}
			reparations = append(reparations, &models.ReparationDetail{
				PayerID:   loserID,
				PayerName: loserName,
				PayeeID:   winnerID,
				PayeeName: winnerName,
				Amount:    paymentPerWinner,
			})
		}
	}

	return reparations
}

func CheckSurrenderSuggestion(state *models.GameState, war *models.AllianceWar) (string, bool) {
	attackerAlliance := FindAllianceByID(state, war.AttackerAllianceID)
	defenderAlliance := FindAllianceByID(state, war.DefenderAllianceID)
	if attackerAlliance == nil || defenderAlliance == nil {
		return "", false
	}

	attackerAssets := calculateAllianceTotalAssets(state, attackerAlliance)
	defenderAssets := calculateAllianceTotalAssets(state, defenderAlliance)

	if attackerAssets > 0 && defenderAssets < attackerAssets*0.30 {
		return war.DefenderAllianceID, true
	}
	if defenderAssets > 0 && attackerAssets < defenderAssets*0.30 {
		return war.AttackerAllianceID, true
	}
	return "", false
}

func calculateAllianceTotalAssets(state *models.GameState, alliance *models.Alliance) float64 {
	total := 0.0
	for _, memberID := range alliance.MemberIDs {
		player := findPlayerInState(state, memberID)
		if player != nil && !player.IsBankrupt && !player.IsDefeated {
			total += CalculateNetWorth(player, state.Exchanges)
		}
	}
	return total
}

func CreateSanctionProposal(state *models.GameState, initiatorID string, targetID string) (*models.SanctionProposal, error) {
	if initiatorID == targetID {
		return nil, fmt.Errorf("不能对自己发起制裁")
	}

	for _, s := range state.ActiveSanctions {
		if s.TargetID == targetID {
			return nil, fmt.Errorf("该玩家已被制裁中，不能重复制裁")
		}
	}

	for _, sp := range state.SanctionProposals {
		if sp.TargetID == targetID && sp.Status == models.SanctionProposalPending {
			return nil, fmt.Errorf("该玩家已有待附议的制裁提案")
		}
	}

	proposal := &models.SanctionProposal{
		ID:          fmt.Sprintf("sanction-proposal-%d-%s", state.Turn, initiatorID[:8]),
		InitiatorID: initiatorID,
		TargetID:    targetID,
		SeconderIDs: make([]string, 0),
		Status:      models.SanctionProposalPending,
		CreatedTurn: state.Turn,
		ExpiryTurn:  state.Turn + SanctionProposalExpiry,
	}

	state.SanctionProposals = append(state.SanctionProposals, proposal)
	return proposal, nil
}

func SecondSanctionProposal(state *models.GameState, seconderID string, proposalID string) error {
	var proposal *models.SanctionProposal
	for _, sp := range state.SanctionProposals {
		if sp.ID == proposalID {
			proposal = sp
			break
		}
	}
	if proposal == nil {
		return fmt.Errorf("制裁提案不存在")
	}
	if proposal.Status != models.SanctionProposalPending {
		return fmt.Errorf("制裁提案不在待附议状态")
	}
	if proposal.InitiatorID == seconderID {
		return fmt.Errorf("发起人不能附议自己的提案")
	}
	if proposal.TargetID == seconderID {
		return fmt.Errorf("被制裁目标不能附议")
	}
	for _, sid := range proposal.SeconderIDs {
		if sid == seconderID {
			return fmt.Errorf("你已经附议过该提案")
		}
	}

	proposal.SeconderIDs = append(proposal.SeconderIDs, seconderID)

	ModifyDiplomacyValue(state, seconderID, proposal.TargetID, -10.0, "附议制裁提案")

	if len(proposal.SeconderIDs) >= SanctionRequiredSeconders {
		proposal.Status = models.SanctionProposalActive
		activateSanction(state, proposal)
	}

	return nil
}

func activateSanction(state *models.GameState, proposal *models.SanctionProposal) {
	sanction := &models.Sanction{
		ID:          fmt.Sprintf("sanction-%d-%s", state.Turn, proposal.TargetID[:8]),
		TargetID:    proposal.TargetID,
		InitiatorID: proposal.InitiatorID,
		CreatedTurn: state.Turn,
		ExpiryTurn:  state.Turn + SanctionDuration,
		TurnsLeft:   SanctionDuration,
	}
	state.ActiveSanctions = append(state.ActiveSanctions, sanction)

	ModifyDiplomacyValue(state, proposal.InitiatorID, proposal.TargetID, -10.0, "发起制裁")
}

func ProcessExpiredSanctionProposals(state *models.GameState) {
	for _, sp := range state.SanctionProposals {
		if sp.Status != models.SanctionProposalPending {
			continue
		}
		if state.Turn >= sp.ExpiryTurn {
			sp.Status = models.SanctionProposalExpired
		}
	}
}

func ProcessActiveSanctions(state *models.GameState) []*models.SanctionEvent {
	events := make([]*models.SanctionEvent, 0)

	active := make([]*models.Sanction, 0, len(state.ActiveSanctions))
	for _, s := range state.ActiveSanctions {
		s.TurnsLeft--
		if s.TurnsLeft <= 0 {
			events = append(events, &models.SanctionEvent{
				SanctionID: s.ID,
				EventType:  "sanction_expired",
				TargetID:   s.TargetID,
				InitiatorID: s.InitiatorID,
				Turn:       state.Turn,
			})
			continue
		}

		player := findPlayerInState(state, s.TargetID)
		if player != nil && !player.IsBankrupt && !player.IsDefeated {
			fee := player.Credits * SanctionMaintenanceRate
			if fee > 0 {
				player.Credits -= fee
				events = append(events, &models.SanctionEvent{
					SanctionID:    s.ID,
					EventType:     "sanction_maintenance",
					TargetID:      s.TargetID,
					TargetName:    player.Name,
					InitiatorID:   s.InitiatorID,
					Turn:          state.Turn,
					MaintenanceFee: fee,
				})
			}
		}

		active = append(active, s)
	}
	state.ActiveSanctions = active

	return events
}

func IsPlayerSanctioned(state *models.GameState, playerID string) bool {
	for _, s := range state.ActiveSanctions {
		if s.TargetID == playerID {
			return true
		}
	}
	return false
}

func GetSanctionFeeMultiplier(state *models.GameState, playerID string) float64 {
	if IsPlayerSanctioned(state, playerID) {
		return SanctionFeeMultiplier
	}
	return 1.0
}

func CanPlayerJoinAlliance(state *models.GameState, playerID string) bool {
	return !IsPlayerSanctioned(state, playerID)
}
