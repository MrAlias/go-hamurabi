package main

import (
	"fmt"
	"math"
	"math/rand"
)

func (s *gameSession) checkForPlague() bool {
	if s.state.year > 0 && rand.Intn(15) == 0 {
		fmt.Println("\nA horrible plague has struck! Many have died!")
		s.state.died = s.state.population / (rand.Intn(4) + 2)
		s.state.cows = s.state.cows / 4
		s.state.population -= s.state.died
		s.totalDead += s.state.died
		return true
	}
	return false
}

func (s *gameSession) printYearResults() {
	var plague bool
	var palaceComplete bool
	if s.state.year > 0 {
		plague, palaceComplete = s.doNumbers()
	}
	fmt.Printf("\nMy lord, in the year %d, I beg to report to you that %d people starved, %d were born, and %d "+
		"came to the city.\n", s.state.year, s.state.starved, s.state.born, s.state.migrated)
	fmt.Printf("Population is now %d.\n", s.state.population)
	fmt.Printf("The city owns %d acres of land, and has %d granaries.\n", s.state.acres, s.state.granary)
	if palaceComplete {
		fmt.Println("My Lord, your workers have completed work on your palace!")
	}
	switch {
	case s.state.palace3:
		fmt.Println("You are residing in a massive bustling palace, together with you family, many retainers, royal merchants, and visiting diplomats.")
	case s.state.palace2:
		fmt.Println("You are residing in a huge palace, together with your family and many retainers.")
	case s.state.palace1:
		fmt.Println("You are residing in a large palace, together with your family and closest retainers.")
	}

	if s.state.buildingPalace > -1 {
		switch {
		case plague:
			fmt.Println("My Dread Lord, I regret that due to the plague, no work was done on your palace!")
			fallthrough
		case !s.state.palace1:
			fmt.Printf("Construction on your palace is underway, and will be completed in %d years.\n", 5-s.state.buildingPalace)
		case s.state.palace1:
			fmt.Printf("Expansion of your palace is underway, and will be completed in %d years.\n", 5-s.state.buildingPalace)
		case s.state.palace2:
			fmt.Printf("Expansion of your palace is underway, and will be completed in %d years.\n", 5-s.state.buildingPalace)
		}
	}

	// we can't support the cows - so they are killed
	if s.state.forceSlaughtered > 0 {
		fmt.Printf("As we lacked the land to support them, %d cows were slaughtered!\n", s.state.forceSlaughtered)
	}
	fmt.Printf("The city keeps %d cows whose product fed %d people this year.\n", s.state.cows, s.state.cowsFed)

	if s.state.acres < 1 || s.state.planted == 0 {
		fmt.Printf("Traders report that %s harvested %d bushels per acrs.\n", s.otherCityStates[rand.Intn(len(s.otherCityStates)-1)], s.state.bYield)
	} else {
		fmt.Printf("We have harvested %d bushels per acre.\n", s.state.bYield)
	}

	if s.state.nonFarmer > 0 && s.state.tradeGoods > 0 {
		fmt.Printf("Thanks to having %d citizens not required to farm, trade goods and vegatables brought in %d "+
			"bushels of grain.\n", s.state.nonFarmer, s.state.tradeGoods)
	}

	fmt.Printf("Rats ate %d bushels of grain.\n", s.state.pests)
	fmt.Printf("We now have %d bushels in store.\n", s.state.bushels)
	fmt.Printf("We have distributed a total of %d hand plows amongst the people.\n", s.state.plows)
	fmt.Printf("Land is trading at %d bushels per acre.\n", s.state.tradeVal)
	s.state.year += 1
}

func (s *gameSession) doNumbers() (bool, bool) {
	plague := s.checkForPlague()
	var palaceComplete bool

	if !plague && s.state.buildingPalace > -1 {
		s.state.buildingPalace++
	}
	if s.state.buildingPalace > 4 {
		switch s.palaceBuilding {
		case 1:
			s.state.palace1 = true
		case 2:
			s.state.palace2 = true
		case 3:
			s.state.palace3 = true
		}
		palaceComplete = true
		s.state.buildingPalace = -1
		s.palaceBuilding = -1
	}

	s.state.tradeVal = 17 + rand.Intn(10)
	s.state.bYield = rand.Intn(9) + 1
	// cows
	s.doCows()
	// starvation & population
	s.doStarvation(plague)
	s.checkForOverthrow()

	s.state.population += s.state.born
	s.avgStarved = int(float64(s.state.starved) / float64(s.state.population) * 100)
	s.state.population -= s.state.starved // children die too
	// migration
	s.doMigration(plague)
	// pests
	s.doPests()
	// agricultural results
	s.doAgriculture()

	s.state.tradeGoods = s.state.nonFarmer * (rand.Intn(49) + 1)
	s.state.bushels += s.state.tradeGoods
	s.totalDead += s.state.starved
	s.avgPestEaten += s.state.pests
	s.avgBushelsAvail += s.state.bushels
	return plague, palaceComplete
}

func (s *gameSession) doAgriculture() {
	s.state.bushels += (s.state.planted - s.state.cows*3) * s.state.bYield
	s.state.bushels -= s.state.pests
	if s.state.bushels < 0 {
		s.state.bushels = 0
	}

	// although the peasants don't have to sow, land must be tended or it will become wasted and be reclaimed by nature
	// some lands are tended by the royal staff, and although they can be sold, they CAN'T go to waste
	royalLands := 500
	fieldMaintPerPop := 30
	maxAcresMaint := s.state.population * fieldMaintPerPop
	// we don't lose the royal-held lands to wastage from lack of peasants
	if s.state.acres > royalLands {
		// if there aren't enough peasants to maintain our acreage
		if maxAcresMaint < s.state.acres {
			s.state.acresWastage = int(math.Abs(float64(maxAcresMaint - (s.state.acres - royalLands))))
			fmt.Printf("Due to a lack of peasants to work the land, %d acres have wasted and are lost!\n", s.state.acresWastage)
		} else {
			s.state.acresWastage = 0
		}
	} else {
		s.state.acresWastage = 0
	}
	s.state.acres -= s.state.acresWastage
	s.totAcresWasted += s.state.acresWastage
	if s.state.acres < royalLands {
		s.state.acres = royalLands
		fmt.Println("However your personal retainers protected your personal estate!")
	}
}

func (s *gameSession) doPests() {
	granaryProtectMultiplier := 3000
	unprotectedGrain := s.state.bushels - s.state.granary*granaryProtectMultiplier
	if unprotectedGrain < 0 {
		unprotectedGrain = 0
	}
	s.state.pests = int(float64(unprotectedGrain) / float64(rand.Intn(4)+3))
}

func (s *gameSession) doMigration(plague bool) {
	var cowMigrantAttraction int
	switch {
	case s.state.cows <= 3:
		cowMigrantAttraction = 0
	case s.state.cows > 3 && s.state.population <= 500:
		cowMigrantAttraction = s.state.cows * 5
	case s.state.cows > 3 && s.state.population <= 10000:
		cowMigrantAttraction = s.state.cows * 3
	case s.state.cows > 3 && s.state.population > 10000:
	default:
		cowMigrantAttraction = 0
	}
	if plague {
		// people don't come to a place with a plague
		s.state.migrated = (int(0.1*float64(rand.Intn(s.state.population)+1)) + cowMigrantAttraction) / 5
	} else {
		s.state.migrated = int(0.1*float64(rand.Intn(s.state.population)+1)) + cowMigrantAttraction
	}
	s.state.population += s.state.migrated
}

func (s *gameSession) doStarvation(plague bool) {
	s.state.starved = s.state.population - (s.state.popFed + s.state.cows*s.state.cowMultiplier)
	if s.state.starved < 0 {
		s.state.starved = 0
	}
	s.avgStarved = int(float64(s.state.starved) / float64(s.state.population) * 100)
	s.state.born = int(float64(s.state.population) / float64(rand.Intn(8)+2))
	if plague {
		s.state.born /= 2 // children die from the plague as well
	}
}

func (s *gameSession) doCows() {
	if s.state.acres < s.state.cows*3 {
		s.state.forceSlaughtered = 0
		if s.state.acres <= 2 {
			s.state.forceSlaughtered = s.state.cows
		} else {
			s.state.forceSlaughtered = (s.state.acres / 3) % s.state.cows
		}
		s.state.cows -= s.state.forceSlaughtered
	}
	if s.state.cows*s.state.cowMultiplier > s.state.population {
		s.state.cowsFed = s.state.population
	} else {
		s.state.cowsFed = s.state.cows * s.state.cowMultiplier
	}
}

func (s *gameSession) checkForOverthrow() {
	if s.state.starved > int(0.45*float64(s.state.population)) {
		fmt.Printf("\nYou starved %d out of your population of only %d, this has caused you to be deposed by force!\n",
			s.state.starved, s.state.population)
		s.totalDead += s.state.starved
		s.endOfReign()
	}

	if s.state.population < 10 {
		fmt.Printf("\nYour continued mismanagement caused your population to decline to the point that the " +
			"remaining peasants fled your land\nYou are left ruling an empty city, as your royal guards and staff escape.\n")
		s.state.population = 0
		s.endOfReign()
	}
}
