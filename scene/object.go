package scene

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/shapes"
)

type Texture struct {
	Size struct{ X, Y int }
	Data []uint8
}

type SceneSpace uint8

const (
	WorldSpace SceneSpace = iota
	ScreenSpace
)

type SceneObject struct {
	Id int32

	Material   *rendering.Material
	Physics    *Physics
	Size       compute.Size
	Shape      shapes.Shape
	Space      SceneSpace
	Controller SceneObjectController
}

type SceneObjectController struct {
	Init   func(self *Node)
	Update func(self *Node, deltaTime time.Duration)
	Input  func(self *Node, event *pb.InputEvent)
}

type SceneObjectArgs struct {
	Material  *rendering.Material
	Physics   *Physics
	Shape     shapes.Shape
	UIElement bool
	Init      func(self *Node)
	Update    func(self *Node, deltaTime time.Duration)
	Input     func(self *Node, event *pb.InputEvent)
}

func NewObject(args SceneObjectArgs) *SceneObject {
	o := &SceneObject{
		Id:       rand.Int32(),
		Material: args.Material,
		Shape:    args.Shape,
		Physics:  args.Physics,
		Controller: SceneObjectController{
			Init:   args.Init,
			Update: args.Update,
			Input:  args.Input,
		},
	}

	if args.UIElement {
		o.Space = ScreenSpace
	} else {
		o.Space = WorldSpace
	}

	if o.Material != nil {
		o.Size = compute.Point{
			X: float64(args.Material.Diffuse.Width),
			Y: float64(args.Material.Diffuse.Height),
		}.Normalize()
	}

	return o
}

func (o *SceneObject) String() string {
	return fmt.Sprintf("id=%d space=%v w=%.2f h=%.2f", o.Id, o.Space, o.Size.X, o.Size.Y)
}
