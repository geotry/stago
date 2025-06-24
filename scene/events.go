package scene

type EventType = uint8

const (
	// Objects
	ObjectSpawned EventType = iota
	ObjectGone
	ObjectMuted
	ObjectPositionChanged
	ObjectScaleChanged
	ObjectTextureChanged
	ObjectAnimationKeyframe

	// Lights
	// ScreenFilter
)

// type ObjectSpawnedEvent struct {
// 	event ObjectSpawned
// }

func NewObjectSpawned(id int) struct {
	Event EventType
	Id    int
} {

	return struct {
		Event EventType
		Id    int
	}{
		Event: ObjectSpawned,
		Id:    id,
	}
}
