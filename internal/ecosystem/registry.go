package ecosystem

var all []Ecosystem

var byPM map[PackageManager]Ecosystem

func init() {
	all = []Ecosystem{
		&npmEcosystem{},
		&yarnEcosystem{},
		&pnpmEcosystem{},
		&bunEcosystem{},
	}

	byPM = make(map[PackageManager]Ecosystem, len(all))
	for _, e := range all {
		byPM[e.Name()] = e
	}
}

// All returns every registered ecosystem.
func All() []Ecosystem {
	return all
}

// ForPM returns the ecosystem for a given package manager, or nil if unknown.
func ForPM(pm PackageManager) Ecosystem {
	return byPM[pm]
}
