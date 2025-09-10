package main

// StationManager is the mediator that manages the trains arriving and departing from the station
type StationManager struct {
	isPlatformFree bool
	trainQueue     []Train
}

// newStationManger creates a new instance of StationManager
func newStationManger() *StationManager {
	return &StationManager{
		isPlatformFree: true,
	}
}

// canArrive checks if a train can arrive at the station
func (s *StationManager) canArrive(t Train) bool {
	if s.isPlatformFree {
		s.isPlatformFree = false
		return true
	}
	s.trainQueue = append(s.trainQueue, t)
	return false
}

// notifyAboutDeparture notifies the station manager about a train's departure
func (s *StationManager) notifyAboutDeparture() {
	if !s.isPlatformFree {
		s.isPlatformFree = true
	}
	if len(s.trainQueue) > 0 {
		firstTrainInQueue := s.trainQueue[0]
		s.trainQueue = s.trainQueue[1:]
		firstTrainInQueue.permitArrival()
	}
}
