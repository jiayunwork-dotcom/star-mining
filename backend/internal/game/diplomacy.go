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
	if HasActiveTradeAgreement(state, buyerID, sellerID) {
		return 0.5
	}
	return 1.0
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

func ProcessDiplomacyTurn(state *models.GameState) []*models.DiplomacyChange {
	changes := make([]*models.DiplomacyChange, 0)

	for _, alliance := range state.Alliances {
		if alliance.Status != models.AllianceActive {
			continue
		}
		for i := 0; i < len(alliance.MemberIDs); i++ {
			for j := i + 1; j < len(alliance.MemberIDs); j++ {
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

	ProcessExpiredInvites(state)

	HandleBankruptMembers(state)

	CleanCooldowns(state)

	return changes
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
