package whip

func CleanUp() {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	for _, resource := range resourceMap {
		resource.Disconnect <- struct{}{}
	}
}
