package app

import (
	"log"
	"strings"
	"sync"

	"github.com/ZenPrivacy/zen-core/filter"
)

const myRulesFilterName = "My rules"

func (a *App) initFilter(filter *filter.Filter) {
	var wg sync.WaitGroup
	for _, filterList := range a.config.GetFilterLists() {
		if !filterList.Enabled {
			continue
		}
		wg.Go(func() {
			contents, err := a.filterListStore.Get(filterList.URL)
			defer contents.Close()
			if err != nil {
				log.Printf("failed to get filter list %q from store: %v", filterList.URL, err)
				return
			}
			if err := filter.AddList(filterList.Name, filterList.Trusted, contents); err != nil {
				log.Printf("failed to add filter list %q to filter: %v", filterList.URL, err)
				return
			}
		})
	}

	wg.Go(func() {
		myRules := a.config.GetRules()
		reader := strings.NewReader(strings.Join(myRules, "\n"))
		if err := filter.AddList(myRulesFilterName, true, reader); err != nil {
			log.Printf("failed to add my rules to filter: %v", err)
			return
		}
	})

	wg.Wait()
}
