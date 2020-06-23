package worker

func Start() {
	go xlsxWorker()
	go autoFetch()
}