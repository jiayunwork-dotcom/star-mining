package game

import (
	"fmt"
	"math/rand"
	"sort"

	"star-mining/internal/models"
)

const (
	MaxSpiesPerPlayer        = 5
	SpyRecruitCost           = 500.0
	SpyMaintenanceJunior     = 50.0
	SpyMaintenanceMiddle     = 100.0
	SpyMaintenanceSenior     = 200.0
	SpyExposureGainMin       = 15
	SpyExposureGainMax       = 30
	SpyExposureThreshold     = 80
	SpyCaptureChance         = 0.5
	SpyExposureIdleDecay     = 5
	SpyLevelUpJuniorToMiddle = 3
	SpyLevelUpMiddleToSenior = 5
	IntelDuration            = 5
	IntelMarketBasePrice     = 500.0
	MaxSpyAttacksPerTarget   = 2
	WarSpySuccessBonus       = 0.15
)

func getPlayerSpies(state *models.GameState, playerID string) []*models.Spy {
	spies := make([]*models.Spy, 0)
	for _, s := range state.Spies {
		if s.PlayerID == playerID {
			spies = append(spies, s)
		}
	}
	return spies
}

func countActiveSpies(state *models.GameState, playerID string) int {
	count := 0
	for _, s := range state.Spies {
		if s.PlayerID == playerID {
			count++
		}
	}
	return count
}

func RecruitSpy(state *models.GameState, playerID string) (*models.Spy, error) {
	player := findPlayerInState(state, playerID)
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	if player.IsDefeated || player.IsBankrupt {
		return nil, fmt.Errorf("玩家已出局")
	}

	if countActiveSpies(state, playerID) >= MaxSpiesPerPlayer {
		return nil, fmt.Errorf("间谍数量已达上限(%d名)", MaxSpiesPerPlayer)
	}

	if player.Credits < SpyRecruitCost {
		return nil, fmt.Errorf("资金不足，招募需要%.0f资金", SpyRecruitCost)
	}

	player.Credits -= SpyRecruitCost

	spy := &models.Spy{
		ID:                fmt.Sprintf("spy-%s-%d", playerID, len(state.Spies)),
		PlayerID:          playerID,
		Level:             models.SpyLevelJunior,
		Exposure:          0,
		Status:            models.SpyStatusIdle,
		CompletedMissions: 0,
		TurnRecruited:     state.Turn,
		IsDoubleAgent:     false,
	}

	state.Spies = append(state.Spies, spy)
	return spy, nil
}

func AssignSpyMission(state *models.GameState, spyID string, ownerPlayerID string, targetPlayerID string, missionType models.SpyMissionType, thirdPartyID string) (*models.SpyMission, error) {
	var spy *models.Spy
	for _, s := range state.Spies {
		if s.ID == spyID {
			spy = s
			break
		}
	}
	if spy == nil {
		return nil, fmt.Errorf("间谍不存在")
	}
	if spy.PlayerID != ownerPlayerID {
		return nil, fmt.Errorf("这不是你的间谍")
	}
	if spy.Status != models.SpyStatusIdle {
		return nil, fmt.Errorf("间谍正在执行任务中")
	}

	target := findPlayerInState(state, targetPlayerID)
	if target == nil {
		return nil, fmt.Errorf("目标玩家不存在")
	}
	if target.IsDefeated || target.IsBankrupt {
		return nil, fmt.Errorf("目标玩家已出局")
	}
	if targetPlayerID == ownerPlayerID {
		return nil, fmt.Errorf("不能对自己执行间谍任务")
	}

	if missionType == models.MissionTurncoat && spy.Level != models.SpyLevelSenior {
		return nil, fmt.Errorf("策反任务只有高级间谍能执行")
	}
	if missionType == models.MissionDiploPressure && spy.Level == models.SpyLevelJunior {
		return nil, fmt.Errorf("初级间谍不能执行外交施压任务")
	}

	if missionType == models.MissionDiploPressure {
		if thirdPartyID == "" {
			return nil, fmt.Errorf("外交施压需要指定第三方玩家")
		}
		if thirdPartyID == ownerPlayerID || thirdPartyID == targetPlayerID {
			return nil, fmt.Errorf("第三方玩家无效")
		}
		tp := findPlayerInState(state, thirdPartyID)
		if tp == nil {
			return nil, fmt.Errorf("第三方玩家不存在")
		}
		if AreAllies(state, ownerPlayerID, thirdPartyID) {
			return nil, fmt.Errorf("不能对盟友所在联盟执行外交施压(防止自损)")
		}
	}

	attackCount := 0
	for _, m := range state.SpyMissions {
		if m.TargetPlayerID == targetPlayerID && !m.Resolved && m.TurnSubmitted == state.Turn {
			attackCount++
		}
	}
	if attackCount >= MaxSpyAttacksPerTarget {
		return nil, fmt.Errorf("目标本回合已被%d次间谍任务攻击，无法再派遣", MaxSpyAttacksPerTarget)
	}

	ownerMissionCount := 0
	for _, m := range state.SpyMissions {
		if m.OwnerPlayerID == ownerPlayerID && !m.Resolved && m.TurnSubmitted == state.Turn {
			ownerMissionCount++
		}
	}

	_ = ownerMissionCount

	spy.Status = models.SpyStatusOnMission

	mission := &models.SpyMission{
		ID:             fmt.Sprintf("mission-%s-%d", spyID, state.Turn),
		SpyID:          spyID,
		OwnerPlayerID:  ownerPlayerID,
		TargetPlayerID: targetPlayerID,
		ThirdPartyID:   thirdPartyID,
		MissionType:    missionType,
		TurnSubmitted:  state.Turn,
		Success:        false,
		Result:         "",
		Resolved:       false,
	}

	state.SpyMissions = append(state.SpyMissions, mission)
	return mission, nil
}

func getMissionSuccessRate(spy *models.Spy, missionType models.SpyMissionType, state *models.GameState, ownerPlayerID string, targetPlayerID string) float64 {
	baseRate := 0.0

	switch missionType {
	case models.MissionStealTech:
		switch spy.Level {
		case models.SpyLevelJunior:
			baseRate = 0.30
		case models.SpyLevelMiddle:
			baseRate = 0.50
		case models.SpyLevelSenior:
			baseRate = 0.70
		}
	case models.MissionEconSabotage:
		baseRate = 0.60
	case models.MissionIntelGather:
		switch spy.Level {
		case models.SpyLevelJunior:
			baseRate = 0.50
		case models.SpyLevelMiddle:
			baseRate = 0.70
		case models.SpyLevelSenior:
			baseRate = 0.90
		}
	case models.MissionTurncoat:
		baseRate = 0.40
	case models.MissionDiploPressure:
		switch spy.Level {
		case models.SpyLevelMiddle:
			baseRate = 0.45
		case models.SpyLevelSenior:
			baseRate = 0.65
		}
	}

	if ArePlayersAtWar(state, ownerPlayerID, targetPlayerID) {
		baseRate += WarSpySuccessBonus
	}

	if baseRate > 1.0 {
		baseRate = 1.0
	}

	return baseRate
}

func ProcessEspionageTurn(state *models.GameState, rng *rand.Rand) *models.SpySection {
	section := &models.SpySection{
		MissionResults:       make([]*models.SpyResult, 0),
		CounterSpyResults:    make([]*models.CounterSpyResult, 0),
		ExpiredIntel:         make([]string, 0),
		LevelUpNotifications: make([]*models.SpyNotification, 0),
		Notifications:        make([]*models.SpyNotification, 0),
	}

	missionResults := make(map[string]*models.SpyResult)

	for _, mission := range state.SpyMissions {
		if mission.Resolved {
			continue
		}
		if mission.TurnSubmitted != state.Turn-1 {
			continue
		}

		var spy *models.Spy
		for _, s := range state.Spies {
			if s.ID == mission.SpyID {
				spy = s
				break
			}
		}
		if spy == nil {
			mission.Resolved = true
			mission.Success = false
			mission.Result = "间谍已被移除"
			continue
		}

		result := resolveMission(state, spy, mission, rng, section)
		missionResults[mission.ID] = result
		section.MissionResults = append(section.MissionResults, result)

		exposureGain := SpyExposureGainMin + rng.Intn(SpyExposureGainMax-SpyExposureGainMin+1)
		spy.Exposure += exposureGain
		result.ExposureGain = exposureGain

		if spy.Exposure >= SpyExposureThreshold && rng.Float64() < SpyCaptureChance {
			spy.Status = models.SpyStatusCaptured
			result.Captured = true
			section.Notifications = append(section.Notifications, &models.SpyNotification{
				Type:    "spy_captured",
				Message: fmt.Sprintf("间谍%s在执行%s任务时被捕，已移除", spy.ID[:12], string(mission.MissionType)),
				Turn:    state.Turn,
				SpyID:   spy.ID,
			})
			removeSpy(state, spy.ID)
		} else {
			spy.Status = models.SpyStatusIdle

			if result.Success {
				spy.CompletedMissions++
				oldLevel := spy.Level
				checkSpyLevelUp(spy)
				if spy.Level != oldLevel {
					section.LevelUpNotifications = append(section.LevelUpNotifications, &models.SpyNotification{
						Type:    "spy_level_up",
						Message: fmt.Sprintf("间谍%s升级为%s", spy.ID[:12], string(spy.Level)),
						Turn:    state.Turn,
						SpyID:   spy.ID,
					})
				}
			}
		}

		mission.Resolved = true
		mission.Success = result.Success
		mission.Result = result.Result
	}

	processCounterEspionage(state, section, rng)

	for _, spy := range state.Spies {
		if spy.Status == models.SpyStatusIdle {
			spy.Exposure -= SpyExposureIdleDecay
			if spy.Exposure < 0 {
				spy.Exposure = 0
			}
		}
	}

	processCounterSpyMaintenance(state, section)

	processSpyMaintenance(state, section)

	processIntelExpiry(state, section)

	processIntelMarketExpiry(state)

	return section
}

func resolveMission(state *models.GameState, spy *models.Spy, mission *models.SpyMission, rng *rand.Rand, section *models.SpySection) *models.SpyResult {
	result := &models.SpyResult{
		SpyID:          spy.ID,
		MissionType:    mission.MissionType,
		TargetPlayerID: mission.TargetPlayerID,
		Success:        false,
		Captured:       false,
		ExposureGain:   0,
		Result:         "",
	}

	if ProcessDoubleAgentLogic(state, spy, rng) {
		result.Result = "任务失败"
		result.IsDoubleAgentFail = true
		if spy.DoubleAgentFor != "" {
			section.Notifications = append(section.Notifications, &models.SpyNotification{
				Type:    "double_agent_sabotage",
				Message: fmt.Sprintf("双面间谍%s故意破坏了一次%s任务", spy.ID[:12], string(mission.MissionType)),
				Turn:    state.Turn,
				SpyID:   spy.ID,
				Details: fmt.Sprintf("任务目标: %s", mission.TargetPlayerID),
			})
		}
		return result
	}

	successRate := getMissionSuccessRate(spy, mission.MissionType, state, mission.OwnerPlayerID, mission.TargetPlayerID)
	roll := rng.Float64()

	if roll >= successRate {
		result.Result = "任务失败"
		return result
	}

	target := findPlayerInState(state, mission.TargetPlayerID)
	if target == nil {
		result.Result = "目标玩家不存在"
		return result
	}

	switch mission.MissionType {
	case models.MissionStealTech:
		resolveStealTech(state, spy, mission, result, target)
	case models.MissionEconSabotage:
		resolveEconSabotage(state, spy, mission, result, target, rng)
	case models.MissionIntelGather:
		resolveIntelGather(state, spy, mission, result, target)
	case models.MissionTurncoat:
		resolveTurncoat(state, spy, mission, result, target, rng)
	case models.MissionDiploPressure:
		resolveDiploPressure(state, spy, mission, result, rng)
	}

	return result
}

func resolveStealTech(state *models.GameState, spy *models.Spy, mission *models.SpyMission, result *models.SpyResult, target *models.Player) {
	if target.TechTree == nil || len(target.TechTree.Techs) == 0 {
		result.Result = "目标无科技可偷取"
		return
	}

	owner := findPlayerInState(state, mission.OwnerPlayerID)
	if owner == nil || owner.TechTree == nil {
		result.Result = "任务失败"
		return
	}

	var stealableTechs []models.TechType
	for techType, targetTech := range target.TechTree.Techs {
		if targetTech.Level <= 1 {
			continue
		}
		ownerTech, exists := owner.TechTree.Techs[techType]
		if exists && ownerTech.Level >= targetTech.Level {
			continue
		}
		stealableTechs = append(stealableTechs, techType)
	}

	if len(stealableTechs) == 0 {
		result.Result = "目标没有可偷取的科技(你已拥有或目标科技等级过低)"
		result.Success = false
		return
	}

	chosenTech := stealableTechs[0]
	targetLevel := target.TechTree.Techs[chosenTech].Level

	ownerTech, exists := owner.TechTree.Techs[chosenTech]
	if !exists {
		owner.TechTree.Techs[chosenTech] = &models.TechLevel{
			Type:  chosenTech,
			Level: targetLevel,
		}
	} else {
		ownerTech.Level = targetLevel
	}
	result.Success = true
	result.StolenTech = string(chosenTech)
	result.Result = fmt.Sprintf("成功窃取科技%s，提升至等级%d", string(chosenTech), targetLevel)
}

func resolveEconSabotage(state *models.GameState, spy *models.Spy, mission *models.SpyMission, result *models.SpyResult, target *models.Player, rng *rand.Rand) {
	damageRate := 0.0
	switch spy.Level {
	case models.SpyLevelJunior:
		damageRate = 0.05
	case models.SpyLevelMiddle:
		damageRate = 0.08
	case models.SpyLevelSenior:
		damageRate = 0.12
	}

	damage := target.Credits * damageRate
	target.Credits -= damage
	if target.Credits < 0 {
		damage += target.Credits
		target.Credits = 0
	}

	result.Success = true
	result.DamageAmount = damage
	result.Result = fmt.Sprintf("成功破坏目标资金，造成%.0f损失", damage)
}

func resolveIntelGather(state *models.GameState, spy *models.Spy, mission *models.SpyMission, result *models.SpyResult, target *models.Player) {
	techList := make([]string, 0)
	if target.TechTree != nil {
		for techType, tech := range target.TechTree.Techs {
			if tech.Level > 0 {
				techList = append(techList, fmt.Sprintf("%s(Lv%d)", string(techType), tech.Level))
			}
		}
	}

	allianceNames := make([]string, 0)
	alliance := FindPlayerAlliance(state, target.ID)
	if alliance != nil {
		allianceNames = append(allianceNames, alliance.Name)
	}

	fleetCount := len(target.Fleets)

	intel := &models.Intelligence{
		ID:               fmt.Sprintf("intel-%s-%d", mission.OwnerPlayerID, state.Turn),
		OwnerPlayerID:    mission.OwnerPlayerID,
		SourcePlayerID:   mission.TargetPlayerID,
		SourceSpyID:      spy.ID,
		Content:          fmt.Sprintf("目标%s情报: 资金%.0f, 舰队%d, 科技%d项", target.Name, target.Credits, fleetCount, len(techList)),
		Summary:          fmt.Sprintf("%s情报快照", target.Name),
		TurnAcquired:     state.Turn,
		ExpiryTurn:       state.Turn + IntelDuration,
		IntelType:        "intel_gather",
		Expired:          false,
		TargetCredits:    target.Credits,
		TargetFleetCount: fleetCount,
		TargetTechs:      techList,
		TargetAlliances:  allianceNames,
	}

	state.Intelligences = append(state.Intelligences, intel)

	result.Success = true
	result.IntelGathered = true
	result.Result = fmt.Sprintf("成功获取%s的情报(有效期%d回合)", target.Name, IntelDuration)
}

func resolveTurncoat(state *models.GameState, spy *models.Spy, mission *models.SpyMission, result *models.SpyResult, target *models.Player, rng *rand.Rand) {
	targetSpies := getPlayerSpies(state, mission.TargetPlayerID)
	if len(targetSpies) == 0 {
		result.Result = "目标没有间谍可策反"
		return
	}

	targetSpy := targetSpies[rng.Intn(len(targetSpies))]
	targetSpy.IsDoubleAgent = true
	targetSpy.DoubleAgentFor = mission.OwnerPlayerID

	result.Success = true
	result.TurncoatSpyID = targetSpy.ID
	result.Result = fmt.Sprintf("成功策反目标间谍%s为双面间谍", targetSpy.ID[:12])
}

func resolveDiploPressure(state *models.GameState, spy *models.Spy, mission *models.SpyMission, result *models.SpyResult, rng *rand.Rand) {
	if mission.ThirdPartyID == "" {
		result.Result = "未指定第三方玩家"
		return
	}

	change := ModifyDiplomacyValue(state, mission.TargetPlayerID, mission.ThirdPartyID, -20.0, "间谍外交施压")
	if change == nil {
		result.Result = "外交施压失败(双方处于交战状态，关系值修改被忽略)"
		return
	}

	result.Success = true
	result.DiploTarget = mission.TargetPlayerID
	result.DiploThirdParty = mission.ThirdPartyID
	result.Result = fmt.Sprintf("成功降低目标与第三方的外交关系值%.0f", change.Change)
}

func checkSpyLevelUp(spy *models.Spy) {
	if spy.Level == models.SpyLevelJunior && spy.CompletedMissions >= SpyLevelUpJuniorToMiddle {
		spy.Level = models.SpyLevelMiddle
	} else if spy.Level == models.SpyLevelMiddle && spy.CompletedMissions >= SpyLevelUpJuniorToMiddle+SpyLevelUpMiddleToSenior {
		spy.Level = models.SpyLevelSenior
	}
}

func removeSpy(state *models.GameState, spyID string) {
	active := make([]*models.Spy, 0, len(state.Spies))
	for _, s := range state.Spies {
		if s.ID != spyID {
			active = append(active, s)
		}
	}
	state.Spies = active
}

func processCounterSpyMaintenance(state *models.GameState, section *models.SpySection) {
	for _, player := range state.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		level := GetCounterSpyLevel(state, player.ID)
		cost := GetCounterSpyCost(level)

		if cost <= 0 {
			continue
		}

		if player.Credits >= cost {
			player.Credits -= cost
		} else {
			player.Credits = 0
			SetCounterSpyLevel(state, player.ID, models.CounterSpyLow)
			section.Notifications = append(section.Notifications, &models.SpyNotification{
				Type:    "counter_spy_downgraded",
				Message: fmt.Sprintf("资金不足，反间谍等级自动降至最低"),
				Turn:    state.Turn,
			})
		}
	}
}

func processSpyMaintenance(state *models.GameState, section *models.SpySection) {
	for _, player := range state.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		spies := getPlayerSpies(state, player.ID)
		if len(spies) == 0 {
			continue
		}

		totalMaintenance := 0.0
		for _, spy := range spies {
			switch spy.Level {
			case models.SpyLevelJunior:
				totalMaintenance += SpyMaintenanceJunior
			case models.SpyLevelMiddle:
				totalMaintenance += SpyMaintenanceMiddle
			case models.SpyLevelSenior:
				totalMaintenance += SpyMaintenanceSenior
			}
		}

		if player.Credits >= totalMaintenance {
			player.Credits -= totalMaintenance
			continue
		}

		sort.Slice(spies, func(i, j int) bool {
			return spies[i].Exposure > spies[j].Exposure
		})

		for len(spies) > 0 {
			removed := spies[0]
			spies = spies[1:]

			section.Notifications = append(section.Notifications, &models.SpyNotification{
				Type:    "spy_disbanded",
				Message: fmt.Sprintf("资金不足，间谍%s(暴露值%d)被解散", removed.ID[:12], removed.Exposure),
				Turn:    state.Turn,
				SpyID:   removed.ID,
			})
			removeSpy(state, removed.ID)

			totalMaintenance = 0
			for _, spy := range spies {
				switch spy.Level {
				case models.SpyLevelJunior:
					totalMaintenance += SpyMaintenanceJunior
				case models.SpyLevelMiddle:
					totalMaintenance += SpyMaintenanceMiddle
				case models.SpyLevelSenior:
					totalMaintenance += SpyMaintenanceSenior
				}
			}

			if player.Credits >= totalMaintenance {
				player.Credits -= totalMaintenance
				break
			}
		}
	}
}

func processCounterEspionage(state *models.GameState, section *models.SpySection, rng *rand.Rand) {
	for _, player := range state.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		counterLevel := models.CounterSpyLow
		for _, s := range state.CounterSpySettings {
			if s.PlayerID == player.ID {
				counterLevel = s.Level
				break
			}
		}

		effectiveLevel := counterLevel
		atWar := FindActiveWarForPlayer(state, player.ID) != nil
		if atWar {
			switch counterLevel {
			case models.CounterSpyHigh:
				effectiveLevel = models.CounterSpyMedium
			case models.CounterSpyMedium:
				effectiveLevel = models.CounterSpyLow
			case models.CounterSpyLow:
				effectiveLevel = ""
			}
		}

		detectRate := 0.0
		canIdentify := false
		canCounter := false

		switch effectiveLevel {
		case models.CounterSpyLow:
			detectRate = 0.20
		case models.CounterSpyMedium:
			detectRate = 0.40
			canIdentify = true
		case models.CounterSpyHigh:
			detectRate = 0.60
			canIdentify = true
			canCounter = true
		}

		if detectRate <= 0 {
			continue
		}

		targetMissions := make([]*models.SpyMission, 0)
		for _, m := range state.SpyMissions {
			if m.TargetPlayerID == player.ID && m.TurnSubmitted == state.Turn-1 && m.Resolved {
				targetMissions = append(targetMissions, m)
			}
		}

		for _, mission := range targetMissions {
			if rng.Float64() >= detectRate {
				continue
			}

			csResult := &models.CounterSpyResult{
				Detected:       true,
				Identified:     canIdentify,
				CounterDone:    canCounter,
				TargetPlayerID: player.ID,
			}

			if canIdentify {
				csResult.SourcePlayerID = mission.OwnerPlayerID
			}

			var spy *models.Spy
			for _, s := range state.Spies {
				if s.ID == mission.SpyID {
					spy = s
					break
				}
			}

			if spy != nil {
				csResult.SpyID = spy.ID

				if canCounter {
					spy.Exposure += 30
					csResult.ExposureAdded = 30
				}

				if spy.IsDoubleAgent && spy.DoubleAgentFor != "" && spy.DoubleAgentFor != player.ID {
					removeSpy(state, spy.ID)
					csResult.RemovedSpyID = spy.ID
					section.Notifications = append(section.Notifications, &models.SpyNotification{
						Type:    "double_agent_removed",
						Message: fmt.Sprintf("反间谍检测到双面间谍%s，已移除", spy.ID[:12]),
						Turn:    state.Turn,
						SpyID:   spy.ID,
					})
				}
			}

			section.CounterSpyResults = append(section.CounterSpyResults, csResult)
		}
	}
}

func SetCounterSpyLevel(state *models.GameState, playerID string, level models.CounterSpyLevel) error {
	player := findPlayerInState(state, playerID)
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	cost := 0.0
	switch level {
	case models.CounterSpyLow:
		cost = 100.0
	case models.CounterSpyMedium:
		cost = 250.0
	case models.CounterSpyHigh:
		cost = 500.0
	default:
		cost = 0.0
	}

	if cost > 0 && player.Credits < cost {
		return fmt.Errorf("资金不足，反间谍等级%s需要%.0f资金/回合", string(level), cost)
	}

	found := false
	for _, s := range state.CounterSpySettings {
		if s.PlayerID == playerID {
			s.Level = level
			found = true
			break
		}
	}
	if !found {
		state.CounterSpySettings = append(state.CounterSpySettings, &models.CounterSpySetting{
			PlayerID: playerID,
			Level:    level,
		})
	}

	return nil
}

func GetCounterSpyLevel(state *models.GameState, playerID string) models.CounterSpyLevel {
	for _, s := range state.CounterSpySettings {
		if s.PlayerID == playerID {
			return s.Level
		}
	}
	return models.CounterSpyLow
}

func GetCounterSpyCost(level models.CounterSpyLevel) float64 {
	switch level {
	case models.CounterSpyLow:
		return 100.0
	case models.CounterSpyMedium:
		return 250.0
	case models.CounterSpyHigh:
		return 500.0
	default:
		return 0.0
	}
}

func ListIntelMarket(state *models.GameState) []*models.IntelMarketListing {
	active := make([]*models.IntelMarketListing, 0)
	for _, l := range state.IntelMarketListings {
		if l.Active {
			active = append(active, l)
		}
	}
	return active
}

func ListIntelForPlayer(state *models.GameState, playerID string) []*models.Intelligence {
	result := make([]*models.Intelligence, 0)
	for _, intel := range state.Intelligences {
		if intel.OwnerPlayerID == playerID && !intel.Expired {
			result = append(result, intel)
		}
	}
	return result
}

func SellIntelOnMarket(state *models.GameState, sellerID string, intelID string, price float64) (*models.IntelMarketListing, error) {
	if IsPlayerSanctioned(state, sellerID) {
		return nil, fmt.Errorf("被制裁的玩家不能在情报市场挂单")
	}

	var intel *models.Intelligence
	for _, i := range state.Intelligences {
		if i.ID == intelID && i.OwnerPlayerID == sellerID && !i.Expired {
			intel = i
			break
		}
	}
	if intel == nil {
		return nil, fmt.Errorf("情报不存在或不属于你")
	}

	remainingTurns := intel.ExpiryTurn - state.Turn
	if remainingTurns <= 0 {
		return nil, fmt.Errorf("情报已过期，无法上架")
	}

	if price <= 0 {
		return nil, fmt.Errorf("价格必须大于0")
	}

	listing := &models.IntelMarketListing{
		ID:             fmt.Sprintf("intel-listing-%s-%d", sellerID, state.Turn),
		SellerID:       sellerID,
		IntelID:        intelID,
		Price:          price,
		BasePrice:      IntelMarketBasePrice,
		CreatedTurn:    state.Turn,
		RemainingTurns: remainingTurns,
		Active:         true,
	}

	state.IntelMarketListings = append(state.IntelMarketListings, listing)
	return listing, nil
}

func BuyIntelFromMarket(state *models.GameState, buyerID string, listingID string) (*models.Intelligence, error) {
	var listing *models.IntelMarketListing
	for _, l := range state.IntelMarketListings {
		if l.ID == listingID && l.Active {
			listing = l
			break
		}
	}
	if listing == nil {
		return nil, fmt.Errorf("挂单不存在或已下架")
	}

	if listing.SellerID == buyerID {
		return nil, fmt.Errorf("不能购买自己出售的情报")
	}

	buyer := findPlayerInState(state, buyerID)
	if buyer == nil {
		return nil, fmt.Errorf("买家不存在")
	}

	remainingTurns := 0
	var originalIntel *models.Intelligence
	for _, i := range state.Intelligences {
		if i.ID == listing.IntelID {
			originalIntel = i
			remainingTurns = i.ExpiryTurn - state.Turn
			break
		}
	}
	if remainingTurns <= 0 || originalIntel == nil {
		listing.Active = false
		return nil, fmt.Errorf("情报已过期")
	}

	marketPrice := (float64(remainingTurns) / float64(IntelDuration)) * IntelMarketBasePrice
	finalPrice := listing.Price
	if finalPrice > marketPrice {
		finalPrice = marketPrice
	}

	if buyer.Credits < finalPrice {
		return nil, fmt.Errorf("资金不足，需要%.0f资金", finalPrice)
	}

	seller := findPlayerInState(state, listing.SellerID)
	buyer.Credits -= finalPrice
	if seller != nil {
		seller.Credits += finalPrice
	}

	listing.Active = false

	newIntel := &models.Intelligence{
		ID:               fmt.Sprintf("intel-%s-%d-copy", buyerID, state.Turn),
		OwnerPlayerID:    buyerID,
		SourcePlayerID:   originalIntel.SourcePlayerID,
		SourceSpyID:      "",
		Content:          originalIntel.Content,
		Summary:          originalIntel.Summary,
		TurnAcquired:     state.Turn,
		ExpiryTurn:       originalIntel.ExpiryTurn,
		IntelType:        originalIntel.IntelType,
		Expired:          false,
		TargetCredits:    originalIntel.TargetCredits,
		TargetFleetCount: originalIntel.TargetFleetCount,
		TargetTechs:      originalIntel.TargetTechs,
		TargetAlliances:  originalIntel.TargetAlliances,
	}

	state.Intelligences = append(state.Intelligences, newIntel)
	return newIntel, nil
}

func CancelIntelListing(state *models.GameState, sellerID string, listingID string) error {
	for _, l := range state.IntelMarketListings {
		if l.ID == listingID && l.SellerID == sellerID && l.Active {
			l.Active = false
			return nil
		}
	}
	return fmt.Errorf("挂单不存在或不属于你")
}

func processIntelExpiry(state *models.GameState, section *models.SpySection) {
	for _, intel := range state.Intelligences {
		if !intel.Expired && state.Turn >= intel.ExpiryTurn {
			intel.Expired = true
			section.ExpiredIntel = append(section.ExpiredIntel, intel.ID)
		}
	}
}

func processIntelMarketExpiry(state *models.GameState) {
	for _, l := range state.IntelMarketListings {
		if !l.Active {
			continue
		}
		var intel *models.Intelligence
		for _, i := range state.Intelligences {
			if i.ID == l.IntelID {
				intel = i
				break
			}
		}
		if intel == nil || intel.Expired {
			l.Active = false
		}
	}
}

func ProcessDoubleAgentLogic(state *models.GameState, spy *models.Spy, rng *rand.Rand) bool {
	if !spy.IsDoubleAgent || spy.DoubleAgentFor == "" {
		return false
	}
	return rng.Float64() < 0.30
}

func GetSpyMaintenanceTotal(state *models.GameState, playerID string) float64 {
	total := 0.0
	for _, spy := range state.Spies {
		if spy.PlayerID == playerID {
			switch spy.Level {
			case models.SpyLevelJunior:
				total += SpyMaintenanceJunior
			case models.SpyLevelMiddle:
				total += SpyMaintenanceMiddle
			case models.SpyLevelSenior:
				total += SpyMaintenanceSenior
			}
		}
	}
	return total
}
