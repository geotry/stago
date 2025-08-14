package examples

import (
	"math"

	"github.com/geotry/stago/compute"
	"github.com/geotry/stago/rendering"
	"github.com/geotry/stago/scene"
)

func NewGround(rm *rendering.ResourceManager) *scene.SceneObject {
	shape := compute.NewCube()

	transform := compute.NewTransform(nil)
	transform.Rotation = compute.NewQuaternionFromEuler(compute.Vector3{X: math.Pi / 2})
	transform.Scale.Z = 1

	transform.ObjectToWorld(shape.Geometry, shape.Geometry)
	transform.ObjectToWorld(shape.Collider, shape.Collider)
	transform.ObjectToWorld(shape.Normals, shape.Normals)

	return scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 16, 16, 64, 64),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64_specular.png", rendering.Specular, 16, 16, 64, 64),
			Shininess: 32.0,
			Opaque:    true,
		},
		Shape:   shape,
		Physics: &scene.Physics{},
		Init: func(self *scene.Node) {
			self.IsKinematic = true
		},
	})
}
