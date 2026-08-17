package blocks

import graft "github.com/lrnxzz/graft"

var fullCube = []graft.AABB{
	graft.Box(graft.Vec3(0, 0, 0), graft.Vec3(1, 1, 1)),
}

func Solid(state graft.BlockState) bool {
	block, ok := Of(state)
	if !ok {
		return false
	}

	return block.Solid()
}

func Collision(state graft.BlockState) []graft.AABB {
	if !Solid(state) {
		return nil
	}

	return fullCube
}
