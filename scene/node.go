package scene

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/geotry/rass/compute"
)

type Transform struct {
	Position compute.Vector3
	Rotation compute.Quaternion
	Scale    compute.Vector3
}

type Node struct {
	Id uint32

	Object *SceneObject
	Scene  *Scene
	Parent *Node
	Hidden bool

	SpawnTime time.Time

	// This object is attached to a Camera
	Camera *Camera
	// This object is a light source
	Light Light

	// Physics
	Mass              float64         // also Inertia in kg m²
	GravityVelocity   compute.Vector3 // V=d/t (distance/time) In m/s²
	KinematicVelocity compute.Vector3 // V=d/t (distance/time) In m/s²
	RotationVelocity  compute.Vector3
	TerminalVelocity  float64         // Maximum gravity velocity
	KinematicMomentum compute.Vector3 // p=mv (mass/velocity) kg m/s
	RotationMomentum  compute.Vector3 // p=mv (mass/velocity) kg m/s

	Data map[string]any

	Transform Transform

	model *compute.Matrix4
}

func (n *Node) Destroy() {
	n.Scene.scheduleOldObjectInstance(n)
}

func (n *Node) Size() compute.Size {
	return n.Object.Size
}

func (n *Node) WorldPosition() compute.Point {
	pos := n.Transform.Position
	if n.Parent != nil {
		p := n.Parent.WorldPosition()
		pos.X += p.X
		pos.Y += p.Y
		pos.Z += p.Z
	}
	return pos
}

func (c *Node) Move(x, y, z float64) {
	c.Transform.Position.X += x
	c.Transform.Position.Y += y
	c.Transform.Position.Z += z
}

func (c *Node) MoveAt(pos compute.Point) {
	c.Transform.Position.X = pos.X
	c.Transform.Position.Y = pos.Y
	c.Transform.Position.Z = pos.Z
}

// Move objet toward a destination
func (c *Node) MoveToward(pt compute.Point, s float64) {
	d := c.Transform.Position.DistanceTo(pt)
	ds := math.Min(d, s)

	if ds >= 0 {
		dx := math.Abs(pt.X - c.Transform.Position.X)
		dy := math.Abs(pt.Y - c.Transform.Position.Y)
		dz := math.Abs(pt.Z - c.Transform.Position.Z)

		dvx := (dx / (dx + dy + dz)) * s
		dvy := (dy / (dx + dy + dz)) * s
		dvz := (dz / (dx + dy + dz)) * s

		if c.Transform.Position.X < pt.X {
			c.Transform.Position.X += dvx
		} else {
			c.Transform.Position.X -= dvx
		}
		if c.Transform.Position.Y < pt.Y {
			c.Transform.Position.Y += dvy
		} else {
			c.Transform.Position.Y -= dvy
		}
		if c.Transform.Position.Z < pt.Z {
			c.Transform.Position.Z += dvz
		} else {
			c.Transform.Position.Z -= dvz
		}
	}
}

func (n *Node) WorldRotation() compute.Quaternion {
	r := n.Transform.Rotation
	if n.Parent != nil {
		r = r.Mult(n.Parent.WorldRotation())
	}
	return r
}

func (c *Node) SetRotation(rotation compute.Vector3) {
	c.Transform.Rotation = compute.NewQuaternionFromEuler(rotation)
}

func (c *Node) Rotate(rotation compute.Vector3) {
	q := compute.NewQuaternionFromEuler(rotation)
	c.Transform.Rotation = c.Transform.Rotation.Mult(q)
}

func (c *Node) RotateZ(value float64) {
	q := compute.NewQuaternionFromAngle(compute.Vector3{Z: 1}, value)
	c.Transform.Rotation = c.Transform.Rotation.Mult(q)
}

func (c *Node) RotateX(value float64) {
	q := compute.NewQuaternionFromAngle(compute.Vector3{X: 1}, value)
	c.Transform.Rotation = c.Transform.Rotation.Mult(q)
}

func (c *Node) RotateY(value float64) {
	q := compute.NewQuaternionFromAngle(compute.Vector3{Y: 1}, value)
	c.Transform.Rotation = c.Transform.Rotation.Mult(q)
}

func (n *Node) Resize(x, y, z float64) {
	n.Transform.Scale.X += x
	n.Transform.Scale.Y += y
	n.Transform.Scale.Z += z

	if n.Camera != nil {
		n.Camera.updateProjectionMatrix()
		n.Camera.normalizeLookAt()
	}
}

func (n *Node) ScaleAt(x, y float64) {
	n.Transform.Scale.X = x
	n.Transform.Scale.Y = y

	if n.Camera != nil {
		n.Camera.updateProjectionMatrix()
		n.Camera.normalizeLookAt()
	}
}

func (n *Node) IsDescendant(p *Node) bool {
	if n == p || n.Parent == p {
		return true
	}
	if n.Parent == nil {
		return false
	}
	return n.Parent.IsDescendant(p)
}

func (n *Node) ModelMatrix() compute.Matrix {
	n.model.Reset()
	n.model.Scale(n.Transform.Scale)
	// Move to center of rotation before applying rotation
	n.model.Translate(compute.Vector3{X: -0.5, Y: -0.5, Z: -0.5})
	n.model.Rotate(n.WorldRotation())
	n.model.Translate(compute.Vector3{X: -0.5, Y: -0.5, Z: -0.5}.Opposite())
	n.model.Translate(n.WorldPosition())

	// Note: ScreenSpace should have a dedicated projection matrix
	// that is fixed (defined on client?) with space contained in
	// (0, 0), (1, 1) and looking forward (Z:-1), in orthographic view.

	// ScreenSpace objects should be bound to the current active session
	// meaning that server should not stream other session objects. This
	// could be defined in a separate field to allow shareable UI elements
	// {
	//    Scope: Global / Local
	// }
	return n.model.Out
}

func (n *Node) String() string {
	switch {
	case n.Camera != nil:
		return fmt.Sprintf("camera id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Transform.Position)
	case n.Light != nil:
		return fmt.Sprintf("light id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Transform.Position)
	default:
		return fmt.Sprintf("object id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Transform.Position)
	}
}

// Apply an external force with intensity i (in Newton) on an object toward a direction
func (n *Node) Push(f compute.Vector3, i float64) {
	inertia := n.Mass * n.Scene.gravity.Length() // inertia
	q := i - inertia
	if q < 0 {
		log.Printf("push is too weak for node (requires +%vN)", math.Abs(q))
		return
	}

	// Transform force in acceleration
	a := f.Normalize().Mult(q).Div(n.Mass)

	n.Accelerate(a)
}

// Apply an external force (in Newton) on an object at location l (normalized [0, 1] in object space)
// resulting in a displacement (linear motion) and rotation (torque)
// For example, to apply a 100N force on top-right corner of back face:
// PushLocal({ 0, 0, -1 }, 100, { 0.8, 0.8, 1 })
func (n *Node) PushLocal(f compute.Vector3, i float64, l compute.Vector3) {
	// For now, let's assume n is always a cube of 1x1 with the center of rotation at its center (0.5, 0.5, 0.5)
	// Distance vector from center of rotation
	r := compute.Vector3{X: l.X - 0.5, Y: l.Y - 0.5, Z: l.Z - 0.5}
	// Rotate r with rotation of n
	ru := r.Rotate(n.Transform.Rotation)

	fn := f.Normalize()

	// Moment of inertia for a cube, where center of rotation is at the center: I = ms²/6 (s = side)
	// Note: See https://en.wikipedia.org/wiki/List_of_moments_of_inertia for other shapes
	rI := (n.Mass * math.Pow(1, 2)) / 6
	rq := i - rI
	if rq > 0 {
		// Compute torque for rotational motion (𝜏 = r x F)
		// r = distance from pivot point to point where force is applied
		// θ = angle between r and F
		fq := fn.Mult(rq)
		t := ru.Cross(fq)  // torque
		a := t.Div(n.Mass) // rotation acceleration
		n.AccelerateRotation(a)
	}

	// Compute kinematic force by removing the torque intensity
	tk := ru.Cross(fn.Mult(i))              // Compute torque force with no inertia
	ki := i - tk.Length()                   // intensity of kinematic force
	kI := n.Mass * n.Scene.gravity.Length() // inertia
	kq := ki - kI                           // remaining newton after inertia
	if kq > 0 {
		ka := fn.Mult(kq).Div(n.Mass) // kinematic acceleration
		n.Accelerate(ka)
	}
}

// Apply gravity acceleration on object for duration d
func (n *Node) Fall(d time.Duration) {
	// weight := n.Scene.gravity.Mult(n.Mass)
	a := n.Scene.gravity.Mult(float64(d) / float64(time.Second))
	n.GravityVelocity = n.GravityVelocity.Add(a)

	if n.GravityVelocity.Length() >= n.TerminalVelocity {
		n.GravityVelocity = n.Scene.gravity.Normalize().Mult(n.TerminalVelocity)
	}
	n.UpdateMomentum()
}

// Apply a kinematic acceleration (in m/s)
func (n *Node) Accelerate(a compute.Vector3) {
	n.KinematicVelocity = n.KinematicVelocity.Add(a)
	n.UpdateMomentum()
}

// Apply a rotation acceleration (in m/s)
func (n *Node) AccelerateRotation(a compute.Vector3) {
	n.RotationVelocity = n.RotationVelocity.Add(a)
	n.UpdateMomentum()
}

func (n *Node) UpdateMomentum() {
	n.KinematicMomentum = n.KinematicVelocity.Mult(n.Mass).Add(n.GravityVelocity.Mult(n.Mass))
	n.RotationMomentum = n.RotationVelocity.Mult(n.Mass)
}

// Update kinematic velocity with drag force for a duration d
func (n *Node) Drag(d time.Duration) {
	var friction compute.Vector3
	if n.KinematicVelocity.Length() > 1 {
		f := n.KinematicVelocity.Pow(2.0).Mult(0.5 * 1.2 * 1.05 * 1.0)
		friction = n.KinematicVelocity.Normalize().Opposite().Scale(f).Div(n.Mass).Mult((float64(d) / float64(time.Second)))
	} else {
		f := n.KinematicVelocity.Mult(0.5 * 1.2 * 1.05 * 1.0)
		friction = n.KinematicVelocity.Normalize().Opposite().Scale(f).Div(n.Mass).Mult((float64(d) / float64(time.Second)))
	}
	n.KinematicVelocity = n.KinematicVelocity.Add(friction)

	// Todo: rotation friction
	if n.RotationVelocity.Length() > 1 {
		f := n.RotationVelocity.Pow(2.0).Mult(0.5 * 1.2 * 1.05 * 1.0)
		friction = n.RotationVelocity.Normalize().Opposite().Scale(f).Div(n.Mass).Mult((float64(d) / float64(time.Second)))
	} else {
		f := n.RotationVelocity.Mult(0.5 * 1.2 * 1.05 * 1.0)
		friction = n.RotationVelocity.Normalize().Opposite().Scale(f).Div(n.Mass).Mult((float64(d) / float64(time.Second)))
	}
	n.RotationVelocity = n.RotationVelocity.Add(friction)

	n.UpdateMomentum()
}

// Update the position based on Velocity
func (n *Node) UpdateMotion(d time.Duration) {
	n.Drag(d)
	n.Fall(d)

	// Parabolic motion
	if n.KinematicVelocity.Y > 0 {
		// Small hack to remove the vertical kinematic velocity when gravity is stronger (apex point)
		if n.GravityVelocity.Y+n.KinematicVelocity.Y < 0 {
			n.GravityVelocity.Y = n.GravityVelocity.Y + n.KinematicVelocity.Y
			n.KinematicVelocity.Y = 0
		}
	}

	t := float64(d) / float64(time.Second)

	// Position
	v := n.GravityVelocity.Add(n.KinematicVelocity)
	a := v.Mult(t)
	n.Transform.Position = n.Transform.Position.Add(a)

	// Rotation
	ra := n.RotationVelocity.Mult(t)
	qa := compute.NewQuaternionFromEuler(ra)
	n.Transform.Rotation = n.Transform.Rotation.Mult(qa)
}
