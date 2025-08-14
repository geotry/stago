package examples

import (
	"github.com/geotry/stago/compute"
	"github.com/geotry/stago/rendering"
	"github.com/geotry/stago/scene"
)

func NewWall(rm *rendering.ResourceManager) *scene.SceneObject {
	shape := compute.NewCube()

	transform := compute.NewTransform(nil)
	transform.Scale.Z = .1
	transform.Scale.X = 2

	transform.ObjectToWorld(shape.Geometry, shape.Geometry)
	transform.ObjectToWorld(shape.Collider, shape.Collider)
	transform.ObjectToWorld(shape.Normals, shape.Normals)

	return scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 96, 96, 64, 64),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64_specular.png", rendering.Specular, 96, 96, 64, 64),
			Shininess: 128.0,
			Opaque:    true,
		},
		Shape:   shape,
		Physics: &scene.Physics{},
		Init: func(self *scene.Node) {
			self.IsKinematic = true
		},
	})
}
