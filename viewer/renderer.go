package viewer

import (
	_ "embed"
	"runtime"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/shaders/world.vert
var vertexShader string

//go:embed assets/shaders/world.frag
var fragmentShader string

const uploadsPerFrame = 2

type meshJob struct {
	key      gocraft.ChunkPos
	revision int
	world    *gocraft.World
	column   *gocraft.Column
}

type meshResult struct {
	key      gocraft.ChunkPos
	revision int
	geometry mesh.Geometry
}

type Renderer struct {
	program   *gpu.Program
	tileset   *Tileset
	chunks    map[gocraft.ChunkPos]*gpu.Mesh
	revisions map[gocraft.ChunkPos]int
	pending   map[gocraft.ChunkPos]bool
	jobs      chan meshJob
	results   chan meshResult
}

func NewRenderer(tileset *Tileset) (*Renderer, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	renderer := &Renderer{
		program:   program,
		tileset:   tileset,
		chunks:    map[gocraft.ChunkPos]*gpu.Mesh{},
		revisions: map[gocraft.ChunkPos]int{},
		pending:   map[gocraft.ChunkPos]bool{},
		jobs:      make(chan meshJob, 512),
		results:   make(chan meshResult, 512),
	}
	for range runtime.NumCPU() {
		go renderer.mesher()
	}

	return renderer, nil
}

func (r *Renderer) mesher() {
	for job := range r.jobs {
		r.results <- meshResult{
			key:      job.key,
			revision: job.revision,
			geometry: mesh.Chunk(job.world, job.column, r.tileset),
		}
	}
}

func (r *Renderer) Build(world *gocraft.World) {
	loaded := map[gocraft.ChunkPos]bool{}
	for _, column := range world.Columns() {
		key := column.Pos()
		loaded[key] = true

		revision := column.Revision()
		meshed, ok := r.revisions[key]
		if r.pending[key] || (ok && meshed == revision) {
			continue
		}

		job := meshJob{
			key:      key,
			revision: revision,
			world:    world,
			column:   column,
		}
		select {
		case r.jobs <- job:
			r.pending[key] = true
		default:
		}
	}

	for key, chunk := range r.chunks {
		if !loaded[key] {
			chunk.Delete()
			delete(r.chunks, key)
			delete(r.revisions, key)
		}
	}
}

func (r *Renderer) Collect() {
	for range uploadsPerFrame {
		select {
		case result := <-r.results:
			r.install(result)
		default:
			return
		}
	}
}

func (r *Renderer) Flush() {
	for len(r.pending) > 0 {
		r.install(<-r.results)
	}
}

func (r *Renderer) install(result meshResult) {
	delete(r.pending, result.key)
	if previous := r.chunks[result.key]; previous != nil {
		previous.Delete()
	}

	r.chunks[result.key] = result.geometry.Upload()
	r.revisions[result.key] = result.revision
}

func (r *Renderer) Draw(camera Camera) {
	r.program.Use()
	r.program.Mat4("viewProjection", camera.ViewProjection())
	r.tileset.Atlas().Bind(0)
	r.program.Int("atlas", 0)
	for _, chunk := range r.chunks {
		chunk.Draw()
	}
}

func (r *Renderer) Close() {
	close(r.jobs)
}
