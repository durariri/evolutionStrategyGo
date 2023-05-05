package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"sort"
)

func booth(x, y float64) float64 { // 0
	a := x + 2*y - 7
	b := 2*x + y - 5
	return a*a + b*b
}

func camel(x, y float64) float64 { // 0
	return 2*x*x - 1.05*math.Pow(x, 4) + math.Pow(x, 6)/6 + x*y + y*y
}

func beale(x, y float64) float64 { // 0
	return math.Pow(1.5-x+x*y, 2) + math.Pow(2.25-x+x*math.Pow(y, 2), 2) + math.Pow(2.625-x+x*math.Pow(y, 3), 2)
}

func main() {
	space := []float64{-100.0, 100.0}
	fitnessHistory := make([]float64, 1)
	solutions := make([]*Phenotype, 5)
	es := &EvolutionStrategy{
		function:       booth,
		space:          space,
		gens:           1000,
		population:     5,
		totalChildren:  3,
		nargs:          2,
		sigma:          100,
		s:              1,
		fitnessHistory: fitnessHistory,
		solutions:      solutions,
	}
	es.Optimize()
	img, err := createLineChart(es.fitnessHistory)
	if err != nil {
		fmt.Println(err)
	}
	printImage(img)
	fitnessHistory1 := make([]float64, 1)
	es1 := &EvolutionStrategy{
		function:       camel,
		space:          space,
		gens:           1000,
		population:     5,
		totalChildren:  3,
		nargs:          2,
		sigma:          100,
		s:              1,
		fitnessHistory: fitnessHistory1,
		solutions:      solutions,
	}
	es1.Optimize()
	img, err = createLineChart(es1.fitnessHistory)
	if err != nil {
		fmt.Println(err)
	}
	printImage(img)
	fitnessHistory2 := make([]float64, 1)
	es2 := &EvolutionStrategy{
		function:       beale,
		space:          space,
		gens:           1000,
		population:     5,
		totalChildren:  3,
		nargs:          2,
		sigma:          100,
		s:              1,
		fitnessHistory: fitnessHistory2,
		solutions:      solutions,
	}
	es2.Optimize()
	img, err = createLineChart(es2.fitnessHistory)
	if err != nil {
		fmt.Println(err)
	}
	printImage(img)
}

func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b]1337;File=inline=1:%s\a\n", imgBase64Str)
}

func createLineChart(data []float64) (image.Image, error) {
	// Create a new plot and set its dimensions
	p := plot.New()

	p.X.Label.Text = "Gens"
	p.Y.Label.Text = "Fitness"
	p.Add(plotter.NewGrid())

	// Create a new scatter plot with the data
	scatterData := make(plotter.XYs, len(data))
	for i, val := range data {
		scatterData[i].X = float64(i)
		scatterData[i].Y = val
	}
	s, err := plotter.NewScatter(scatterData)
	if err != nil {
		return nil, err
	}
	s.GlyphStyle.Color = color.RGBA{R: 255, A: 255}
	s.GlyphStyle.Radius = 2
	p.Add(s)

	// Draw the plot to a new image and return it
	canvas := vgimg.New(800, 400)
	p.Draw(draw.New(canvas))
	return canvas.Image(), nil
}

type Phenotype struct {
	vals []float64
}

func (p *Phenotype) makeChild(space []float64, sigma float64) *Phenotype {
	child := make([]float64, len(p.vals))
	for i, mu := range p.vals {
		child[i] = rand.NormFloat64()*sigma + mu
		min, max := space[0], space[1]
		child[i] = math.Max(min, child[i])
		child[i] = math.Min(max, child[i])
	}
	return &Phenotype{child}
}

type EvolutionStrategy struct {
	function       func(x, y float64) float64
	space          []float64
	gens           int
	population     int
	totalChildren  int
	solutions      []*Phenotype
	sigma          float64
	s              int
	nargs          int
	fitnessHistory []float64
}

func (es *EvolutionStrategy) Optimize() {
	es.initSolutions()
	for gen := 0; gen < es.gens; gen++ {
		parents := es.bestSolutions()
		children := es.mate(parents)
		es.solutions = append(parents, children...)
		fitness := es.fitness(es.bestSolution())
		if gen%50 == 0 {
			fmt.Printf("gen %v: fitness %v\n", gen, fitness)
		}
		es.fitnessHistory = append(es.fitnessHistory, fitness)
	}
}

func (es *EvolutionStrategy) initSolutions() {
	for i := 0; i < es.population; i++ {
		vals := make([]float64, es.nargs)
		for j := 0; j < es.nargs; j++ {
			vals[j] = rand.Float64()*(es.space[1]-es.space[0]) + es.space[0]
		}
		es.solutions[i] = &Phenotype{vals}
	}
}

func (es *EvolutionStrategy) mate(parents []*Phenotype) []*Phenotype {
	children := make([]*Phenotype, len(parents)*es.totalChildren)
	for i, parent := range parents {
		for j := 0; j < es.totalChildren; j++ {
			child := parent.makeChild(es.space, es.getSigma())
			children[i*es.totalChildren+j] = child
		}
	}
	return children
}

func (es *EvolutionStrategy) bestSolutions() []*Phenotype {
	amount := es.population / es.totalChildren
	sort.Slice(es.solutions, func(i, j int) bool {
		return es.fitness(es.solutions[i]) < es.fitness(es.solutions[j])
	})
	return es.solutions[:amount]
}

func (es *EvolutionStrategy) bestSolution() *Phenotype {
	best := es.solutions[0]
	for _, solution := range es.solutions {
		if es.fitness(solution) < es.fitness(best) {
			best = solution
		}
	}
	return best
}

func (es *EvolutionStrategy) getSigma() float64 {
	return es.sigma / (float64(es.gens)/float64(es.s) + 1)
}

func (es *EvolutionStrategy) fitness(phenotype *Phenotype) float64 {
	return es.function(phenotype.vals[0], phenotype.vals[1])
}
