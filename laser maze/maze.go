package main

import (
	"fmt"

	"github.com/irifrance/gini"
	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"
)

const (
	MIRROR_RIGHT_UP   = 0
	MIRROR_RIGHT_DOWN = 1
	MIRROR_LEFT_UP    = 2
	MIRROR_LEFT_DOWN  = 3

	INPUT_UP    = 0
	INPUT_RIGHT = 1
	INPUT_DOWN  = 2
	INPUT_LEFT  = 3

	OUTPUT_UP    = 2
	OUTPUT_RIGHT = 3
	OUTPUT_DOWN  = 0
	OUTPUT_LEFT  = 1

	TARGET_UP    = 0
	TARGET_RIGHT = 1
	TARGET_DOWN  = 2
	TARGET_LEFT  = 3

	SPLIT_RIGHT_UP_LEFT_DOWN = 0
	SPLIT_RIGHT_DOWN_LEFT_UP = 1
)

func main() {
	g := gini.New()

	var vars [18 * 25]z.Lit

	var litMirror = func(row, col, mirror int) z.Lit {
		n := mirror
		n += col * 18
		n += row * 90
		return vars[n]
	}

	var litInput = func(row, col, in int) z.Lit {
		n := 4
		n += in
		n += col * 18
		n += row * 90
		return vars[n]
	}

	var litOutput = func(row, col, out int) z.Lit {
		n := 8
		n += out
		n += col * 18
		n += row * 90
		return vars[n]
	}

	var litTarget = func(row, col, out int) z.Lit {
		n := 12
		n += out
		n += col * 18
		n += row * 90
		return vars[n]
	}

	var litSplit = func(row, col, split int) z.Lit {
		n := 16
		n += split
		n += col * 18
		n += row * 90
		return vars[n]
	}

	var ands []z.Lit

	c := logic.NewC()

	var empty = func(row, col int) z.Lit {
		return c.Ands(litMirror(row, col, MIRROR_LEFT_DOWN).Not(),
			litMirror(row, col, MIRROR_LEFT_UP).Not(),
			litMirror(row, col, MIRROR_RIGHT_DOWN).Not(),
			litMirror(row, col, MIRROR_RIGHT_UP).Not(),
			litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP).Not(),
			litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN).Not(),
			litTarget(row, col, TARGET_DOWN).Not(),
			litTarget(row, col, TARGET_LEFT).Not(),
			litTarget(row, col, TARGET_RIGHT).Not(),
			litTarget(row, col, TARGET_UP).Not())
	}

	for i, _ := range vars {
		vars[i] = c.Lit()
	}

	// Para cada posición a lo sumo una de las posiciones de espejo, bifurcador y target
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			r := make([]z.Lit, 0, 5)
			r = append(r, litMirror(row, col, MIRROR_LEFT_DOWN), litMirror(row, col, MIRROR_LEFT_UP), litMirror(row, col, MIRROR_RIGHT_DOWN), litMirror(row, col, MIRROR_RIGHT_UP))
			r = append(r, litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN))
			r = append(r, litTarget(row, col, TARGET_DOWN), litTarget(row, col, TARGET_LEFT), litTarget(row, col, TARGET_RIGHT), litTarget(row, col, TARGET_UP))
			card := c.CardSort(r)
			ands = append(ands, card.Leq(1))
		}
	}

	// Las entradas 0 y 2: a lo sumo una de ellas
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			r := make([]z.Lit, 0, 25)
			r = append(r, litInput(row, col, INPUT_DOWN))
			r = append(r, litInput(row, col, INPUT_UP))
			card := c.CardSort(r)
			ands = append(ands, card.Leq(1))
		}
	}

	// Las entradas 1 y 3: a lo sumo una de ellas
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			r := make([]z.Lit, 0, 25)
			r = append(r, litInput(row, col, INPUT_LEFT))
			r = append(r, litInput(row, col, INPUT_RIGHT))
			card := c.CardSort(r)
			ands = append(ands, card.Leq(1))
		}
	}

	// En el perímetro no puede haber espejos que apunten hacia afuera
	// En el perímetro los rayos de salida son todos falsos
	for col := 0; col < 5; col++ {
		row := 0
		ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_DOWN).Not(),
			litMirror(row, col, MIRROR_LEFT_DOWN).Not(),
			litOutput(row, col, OUTPUT_DOWN).Not()))

		row = 4
		ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_UP).Not(),
			litMirror(row, col, MIRROR_LEFT_UP).Not(),
			litOutput(row, col, OUTPUT_UP).Not()))
	}

	for row := 0; row < 5; row++ {
		col := 0
		ands = append(ands, c.Ands(litMirror(row, col, MIRROR_LEFT_UP).Not(),
			litMirror(row, col, MIRROR_LEFT_DOWN).Not(),
			litOutput(row, col, OUTPUT_LEFT).Not()))

		col = 4
		ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_UP).Not(),
			litMirror(row, col, MIRROR_RIGHT_DOWN).Not(),
			litOutput(row, col, OUTPUT_RIGHT).Not()))
	}

	// Los rayos de entrada de una celda COINCIDEN con la salida de la adyacente
	// celdas interiores
	for row := 1; row < 4; row++ {
		for col := 1; col < 4; col++ {

			// La entrada 0 coincide con la salida 0 de la fila siguiente
			ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
			ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

			// La entrada 1 coincide con la salida 1 de la columna siguiente
			ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
			ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

			// La entrada 2 coincide con la salida 2 de la fila anterior
			ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
			ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

			// La entrada 3 coincide con la salida 3 de la fila anterior
			ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
			ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))
		}
	}

	// celdas exteriores (filas 0 y 4)
	for col := 1; col < 4; col++ {
		row := 0
		// La entrada 3 coincide con la salida 3 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
		ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))

		// La entrada 0 coincide con la salida 0 de la fila siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
		ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

		// La entrada 1 coincide con la salida 1 de la columna siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
		ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

		// La entrada 2 está desactivada
		ands = append(ands, litInput(row, col, INPUT_DOWN).Not())

		row = 4
		// La entrada 3 coincide con la salida 3 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
		ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))

		// La entrada 1 coincide con la salida 1 de la columna siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
		ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

		// La entrada 2 coincide con la salida 2 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
		ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

		// La entrada 0 está desactivada
		ands = append(ands, litInput(row, col, INPUT_UP).Not())
	}

	// celdas exteriores ( columnas 0 y 4)
	for row := 1; row < 4; row++ {
		col := 0
		// La entrada 0 coincide con la salida 0 de la fila siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
		ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

		// La entrada 1 coincide con la salida 1 de la columna siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
		ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

		// La entrada 2 coincide con la salida 2 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
		ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

		// La entrada 3 está desactivada
		if row != 3 {
			ands = append(ands, litInput(row, col, INPUT_LEFT).Not())
		}

		col = 4
		// La entrada 0 coincide con la salida 0 de la fila siguiente
		ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
		ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

		// La entrada 2 coincide con la salida 2 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
		ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

		// La entrada 3 coincide con la salida 3 de la fila anterior
		ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
		ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))

		// La entrada 1 está desactivada
		ands = append(ands, litInput(row, col, INPUT_RIGHT).Not())
	}

	// celdas esquinas
	row := 0
	col := 0

	// entradas 0 y 1
	// La entrada 0 coincide con la salida 0 de la fila siguiente
	ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
	ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

	// La entrada 1 coincide con la salida 1 de la columna siguiente
	ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
	ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

	row = 0
	col = 4

	// entradas 0 y 3
	// La entrada 0 coincide con la salida 0 de la fila siguiente
	ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row+1, col, OUTPUT_DOWN)))
	ands = append(ands, c.Implies(litOutput(row+1, col, OUTPUT_DOWN), litInput(row, col, INPUT_UP)))

	// La entrada 3 coincide con la salida 3 de la fila anterior
	ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
	ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))

	row = 4
	col = 0

	// entradas 1 y 2
	// La entrada 1 coincide con la salida 1 de la columna siguiente
	ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col+1, OUTPUT_LEFT)))
	ands = append(ands, c.Implies(litOutput(row, col+1, OUTPUT_LEFT), litInput(row, col, INPUT_RIGHT)))

	// La entrada 2 coincide con la salida 2 de la fila anterior
	ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
	ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

	row = 4
	col = 4

	// entradas 2 y 3
	// La entrada 2 coincide con la salida 2 de la fila anterior
	ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row-1, col, OUTPUT_UP)))
	ands = append(ands, c.Implies(litOutput(row-1, col, OUTPUT_UP), litInput(row, col, INPUT_DOWN)))

	// La entrada 3 coincide con la salida 3 de la fila anterior
	ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col-1, OUTPUT_RIGHT)))
	ands = append(ands, c.Implies(litOutput(row, col-1, OUTPUT_RIGHT), litInput(row, col, INPUT_LEFT)))

	// Trayectorias de salida de los rayos en función de la entrada y su contenido
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {

			// entrada 0 y espejo 0 -> salida 3
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_UP), litMirror(row, col, MIRROR_RIGHT_UP)), c.Ands(litOutput(row, col, OUTPUT_RIGHT), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 0 y espejo 2 -> salida 1
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_UP), litMirror(row, col, MIRROR_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_LEFT), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_RIGHT).Not())))

			// entrada 0 -> no espejo 1
			ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litMirror(row, col, MIRROR_RIGHT_DOWN).Not()))

			// entrada 0 -> no espejo 3
			ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litMirror(row, col, MIRROR_LEFT_DOWN).Not()))

			// entrada 0 -> no target 1,2,3
			ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), c.Ands(litTarget(row, col, TARGET_RIGHT).Not(), litTarget(row, col, TARGET_LEFT).Not(), litTarget(row, col, TARGET_DOWN).Not())))

			// entrada 1 y espejo 1 -> salida 0
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_RIGHT), litMirror(row, col, MIRROR_RIGHT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_RIGHT).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 1 y espejo 0 -> salida 2
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_RIGHT), litMirror(row, col, MIRROR_RIGHT_UP)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_RIGHT).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 1 -> no espejo 2
			ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litMirror(row, col, MIRROR_LEFT_DOWN).Not()))

			// entrada 1 -> no espejo 3
			ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litMirror(row, col, MIRROR_LEFT_UP).Not()))

			// entrada 1 -> no target 0,2,3
			ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), c.Ands(litTarget(row, col, TARGET_UP).Not(), litTarget(row, col, TARGET_LEFT).Not(), litTarget(row, col, TARGET_DOWN).Not())))

			// entrada 2 y espejo 1 -> salida 3
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_DOWN), litMirror(row, col, MIRROR_RIGHT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_RIGHT), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 2 y espejo 3 -> salida 1
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_DOWN), litMirror(row, col, MIRROR_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_LEFT), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_RIGHT).Not())))

			// entrada 2 -> no espejo 0
			ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litMirror(row, col, MIRROR_LEFT_UP).Not()))

			// entrada 2 -> no espejo 2
			ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litMirror(row, col, MIRROR_RIGHT_UP).Not()))

			// entrada 2 -> no target 0,1,3
			ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), c.Ands(litTarget(row, col, TARGET_UP).Not(), litTarget(row, col, TARGET_LEFT).Not(), litTarget(row, col, TARGET_RIGHT).Not())))

			// entrada 3 y espejo 2 -> salida 2
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_LEFT), litMirror(row, col, MIRROR_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_RIGHT).Not(), litOutput(row, col, OUTPUT_DOWN).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 3 y espejo 3 -> salida 0
			ands = append(ands, c.Implies(c.And(litInput(row, col, INPUT_LEFT), litMirror(row, col, MIRROR_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_UP).Not(), litOutput(row, col, OUTPUT_RIGHT).Not(), litOutput(row, col, OUTPUT_LEFT).Not())))

			// entrada 3 -> no espejo 0
			ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litMirror(row, col, MIRROR_RIGHT_DOWN).Not()))

			// entrada 3 -> no espejo 1
			ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litMirror(row, col, MIRROR_RIGHT_UP).Not()))

			// entrada 3 -> no target 0,1,2
			ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), c.Ands(litTarget(row, col, TARGET_UP).Not(), litTarget(row, col, TARGET_RIGHT).Not(), litTarget(row, col, TARGET_DOWN).Not())))

			// entrada 0 y sin espejo y deflector -> salida 0
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_UP), empty(row, col)), litOutput(row, col, OUTPUT_DOWN)))

			// entrada 1 y sin espejo y deflector -> salida 1
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_RIGHT), empty(row, col)), litOutput(row, col, OUTPUT_LEFT)))

			// entrada 2 y sin espejo y deflector -> salida 2
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_DOWN), empty(row, col)), litOutput(row, col, OUTPUT_UP)))

			// entrada 3 y sin espejo y deflector -> salida 3
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_LEFT), empty(row, col)), litOutput(row, col, OUTPUT_RIGHT)))

			// entrada 0 y split 0 -> salidas 0 y 1
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_UP), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_LEFT))))

			// entrada 0 y split 1 -> salidas 0 y 3
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_UP), litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_RIGHT))))

			// entrada 1 y split 0 -> salidas 0 y 1
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_RIGHT), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_LEFT))))

			// entrada 1 y split 1 -> salidas 1 y 2
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_RIGHT), litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_LEFT))))

			// entrada 2 y split 0 -> salidas 2 y 3
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_DOWN), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_RIGHT))))

			// entrada 2 y split 1 -> salidas 2 y 1
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_DOWN), litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_LEFT))))

			// entrada 3 y split 0 -> salidas 3 y 2
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_LEFT), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN)), c.Ands(litOutput(row, col, OUTPUT_UP), litOutput(row, col, OUTPUT_RIGHT))))

			// entrada 3 y split 1 -> salidas 0 y 3
			ands = append(ands, c.Implies(c.Ands(litInput(row, col, INPUT_LEFT), litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP)), c.Ands(litOutput(row, col, OUTPUT_DOWN), litOutput(row, col, OUTPUT_RIGHT))))
		}
	}

	// en las esquinas no puede haber splits
	for _, row := range []int{0, 4} {
		for _, col := range []int{0, 4} {
			ands = append(ands, litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP).Not())
			ands = append(ands, litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN).Not())
		}
	}

	// destinos: si hay un destino no hay rayo de salida
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			for target := 0; target < 4; target++ {
				for output := 0; output < 4; output++ {
					ands = append(ands, c.Implies(litTarget(row, col, target), litOutput(row, col, output).Not()))
				}
			}
		}
	}

	// no pueden estar activas las entradas y las salidas al mismo tiempo
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			ands = append(ands, c.Implies(litInput(row, col, INPUT_RIGHT), litOutput(row, col, OUTPUT_RIGHT).Not()))
			ands = append(ands, c.Implies(litOutput(row, col, OUTPUT_RIGHT), litInput(row, col, INPUT_RIGHT).Not()))
			ands = append(ands, c.Implies(litInput(row, col, INPUT_UP), litOutput(row, col, OUTPUT_UP).Not()))
			ands = append(ands, c.Implies(litOutput(row, col, OUTPUT_UP), litInput(row, col, INPUT_UP).Not()))
			ands = append(ands, c.Implies(litInput(row, col, INPUT_DOWN), litOutput(row, col, OUTPUT_DOWN).Not()))
			ands = append(ands, c.Implies(litOutput(row, col, OUTPUT_DOWN), litInput(row, col, INPUT_DOWN).Not()))
			ands = append(ands, c.Implies(litInput(row, col, INPUT_LEFT), litOutput(row, col, OUTPUT_LEFT).Not()))
			ands = append(ands, c.Implies(litOutput(row, col, OUTPUT_LEFT), litInput(row, col, INPUT_LEFT).Not()))
		}
	}

	// si hay un espejo tiene que haber inputs
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			for mirror := 0; mirror < 4; mirror++ {
				ands = append(ands, c.Implies(litMirror(row, col, MIRROR_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_DOWN))))
				ands = append(ands, c.Implies(litMirror(row, col, MIRROR_LEFT_UP), c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_UP))))
				ands = append(ands, c.Implies(litMirror(row, col, MIRROR_RIGHT_DOWN), c.Ors(litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_DOWN))))
				ands = append(ands, c.Implies(litMirror(row, col, MIRROR_RIGHT_UP), c.Ors(litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_UP))))
			}
		}
	}

	// si hay un split tiene que haber inputs
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			for split := 0; split < 2; split++ {
				ands = append(ands, c.Implies(litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_UP))))
				ands = append(ands, c.Implies(litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_UP))))
			}
		}
	}

	// salidas en función de entradas
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			ands = append(ands,
				c.Implies(litOutput(row, col, OUTPUT_UP),
					c.Ors(
						c.Ands(empty(row, col), litInput(row, col, INPUT_DOWN)),
						c.Ands(litMirror(row, col, MIRROR_LEFT_UP), litInput(row, col, INPUT_LEFT)),
						c.Ands(litMirror(row, col, MIRROR_RIGHT_UP), litInput(row, col, INPUT_RIGHT)),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), c.Ors(litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_RIGHT))),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_LEFT))))))
			ands = append(ands,
				c.Implies(litOutput(row, col, OUTPUT_DOWN),
					c.Ors(
						c.Ands(empty(row, col), litInput(row, col, INPUT_UP)),
						c.Ands(litMirror(row, col, MIRROR_LEFT_DOWN), litInput(row, col, INPUT_LEFT)),
						c.Ands(litMirror(row, col, MIRROR_RIGHT_DOWN), litInput(row, col, INPUT_RIGHT)),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), c.Ors(litInput(row, col, INPUT_UP), litInput(row, col, INPUT_LEFT))),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_UP), litInput(row, col, INPUT_RIGHT))))))
			ands = append(ands,
				c.Implies(litOutput(row, col, OUTPUT_RIGHT),
					c.Ors(
						c.Ands(empty(row, col), litInput(row, col, INPUT_LEFT)),
						c.Ands(litMirror(row, col, MIRROR_RIGHT_DOWN), litInput(row, col, INPUT_DOWN)),
						c.Ands(litMirror(row, col, MIRROR_RIGHT_UP), litInput(row, col, INPUT_UP)),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), c.Ors(litInput(row, col, INPUT_UP), litInput(row, col, INPUT_LEFT))),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_LEFT))))))
			ands = append(ands,
				c.Implies(litOutput(row, col, OUTPUT_LEFT),
					c.Ors(
						c.Ands(empty(row, col), litInput(row, col, INPUT_RIGHT)),
						c.Ands(litMirror(row, col, MIRROR_LEFT_UP), litInput(row, col, INPUT_UP)),
						c.Ands(litMirror(row, col, MIRROR_LEFT_DOWN), litInput(row, col, INPUT_DOWN)),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), c.Ors(litInput(row, col, INPUT_DOWN), litInput(row, col, INPUT_RIGHT))),
						c.Ands(litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN), c.Ors(litInput(row, col, INPUT_UP), litInput(row, col, INPUT_RIGHT))))))
		}
	}

	// los espejos hay que usarlos
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			ands = append(ands,
				c.Implies(litMirror(row, col, MIRROR_LEFT_DOWN),
					c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_DOWN))))

			ands = append(ands,
				c.Implies(litMirror(row, col, MIRROR_LEFT_UP),
					c.Ors(litInput(row, col, INPUT_LEFT), litInput(row, col, INPUT_UP))))

			ands = append(ands,
				c.Implies(litMirror(row, col, MIRROR_RIGHT_DOWN),
					c.Ors(litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_DOWN))))

			ands = append(ands,
				c.Implies(litMirror(row, col, MIRROR_RIGHT_UP),
					c.Ors(litInput(row, col, INPUT_RIGHT), litInput(row, col, INPUT_UP))))

		}
	}

	// los ciclos cerrados de espejos no son posibles
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_UP),
				litMirror(row+1, col, MIRROR_RIGHT_DOWN),
				litMirror(row, col+1, MIRROR_LEFT_UP),
				litMirror(row+1, col+1, MIRROR_LEFT_DOWN)).Not())
		}
	}

	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_UP),
				litMirror(row+2, col, MIRROR_RIGHT_DOWN),
				litMirror(row, col+2, MIRROR_LEFT_UP),
				litMirror(row+2, col+2, MIRROR_LEFT_DOWN)).Not())
		}
	}

	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			ands = append(ands, c.Ands(litMirror(row, col, MIRROR_RIGHT_UP),
				litMirror(row+3, col, MIRROR_RIGHT_DOWN),
				litMirror(row, col+3, MIRROR_LEFT_UP),
				litMirror(row+3, col+3, MIRROR_LEFT_DOWN)).Not())
		}
	}

	ands = append(ands, c.Ands(litMirror(0, 0, MIRROR_RIGHT_UP),
		litMirror(4, 0, MIRROR_RIGHT_DOWN),
		litMirror(0, 4, MIRROR_LEFT_UP),
		litMirror(4, 0, MIRROR_LEFT_DOWN)).Not())

	//----------------------------------------------------------------------------------------------
	// Maze 39
	// ands = append(ands, c.Ands(litInput(3, 0, INPUT_LEFT),
	// 	litInput(2, 3, INPUT_RIGHT),
	// 	litInput(0, 4, INPUT_UP),
	// 	litMirror(4, 0, MIRROR_RIGHT_DOWN),
	// 	litMirror(0, 2, MIRROR_LEFT_UP)))

	// (Assume) Target
	// for row := 0; row < 5; row++ {
	// 	for col := 0; col < 5; col++ {
	// 		for target := 0; target < 4; target++ {
	// 			// ubicación de los targets
	// 			if (row == 2 && col == 3 && target == TARGET_RIGHT) ||
	// 				(row == 0 && col == 4 && target == TARGET_UP) {
	// 				ands = append(ands, c.Ands(litTarget(row, col, target)))
	// 			} else {
	// 				ands = append(ands, c.Ands(litTarget(row, col, target).Not()))
	// 			}
	// 		}
	// 	}
	// }

	// 5 espejos en total
	// r := make([]z.Lit, 0, 5)
	// for row := 0; row < 5; row++ {
	// 	for col := 0; col < 5; col++ {
	// 		r = append(r, litMirror(row, col, MIRROR_LEFT_DOWN), litMirror(row, col, MIRROR_LEFT_UP), litMirror(row, col, MIRROR_RIGHT_DOWN), litMirror(row, col, MIRROR_RIGHT_UP))
	// 	}
	// }
	// card := c.CardSort(r)
	// ands = append(ands, card.Leq(5))
	// ands = append(ands, card.Geq(5))

	// 1 split
	// r = make([]z.Lit, 0, 5)
	// for row := 0; row < 5; row++ {
	// 	for col := 0; col < 5; col++ {
	// 		r = append(r, litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN))
	// 	}
	// }
	// card = c.CardSort(r)
	// ands = append(ands, card.Leq(1))
	// ands = append(ands, card.Geq(1))

	//----------------------------------------------------------------------------------------------

	//----------------------------------------------------------------------------------------------
	// Maze 37
	ands = append(ands, c.Ands(litInput(3, 0, INPUT_LEFT),
		litInput(4, 1, INPUT_DOWN),
		litInput(2, 0, INPUT_RIGHT),
		litSplit(1, 3, SPLIT_RIGHT_DOWN_LEFT_UP)))

	// (Assume) Target
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			for target := 0; target < 4; target++ {
				// ubicación de los targets
				if (row == 4 && col == 1 && target == TARGET_DOWN) ||
					(row == 2 && col == 0 && target == TARGET_RIGHT) {
					ands = append(ands, c.Ands(litTarget(row, col, target)))
				} else {
					ands = append(ands, c.Ands(litTarget(row, col, target).Not()))
				}
			}
		}
	}

	// 4 espejos en total
	r := make([]z.Lit, 0, 5)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			r = append(r, litMirror(row, col, MIRROR_LEFT_DOWN), litMirror(row, col, MIRROR_LEFT_UP), litMirror(row, col, MIRROR_RIGHT_DOWN), litMirror(row, col, MIRROR_RIGHT_UP))
		}
	}
	card := c.CardSort(r)
	ands = append(ands, card.Leq(4))
	ands = append(ands, card.Geq(4))

	// TODO: 1 split
	r = make([]z.Lit, 0, 5)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			r = append(r, litSplit(row, col, SPLIT_RIGHT_DOWN_LEFT_UP), litSplit(row, col, SPLIT_RIGHT_UP_LEFT_DOWN))
		}
	}
	card = c.CardSort(r)
	ands = append(ands, card.Leq(1))
	ands = append(ands, card.Geq(1))

	//----------------------------------------------------------------------------------------------

	f := c.Ands(ands...)
	c.ToCnfFrom(g, f)
	g.Assume(f)

	if g.Solve() == 1 {
		for row := 0; row < 5; row++ {
			for col := 0; col < 5; col++ {
				for mirror := 0; mirror < 4; mirror++ {
					if g.Value(litMirror(row, col, mirror)) {
						fmt.Printf("mirror %d at (%d,%d)\n", mirror, row, col)
					}
				}
			}
		}

		for row := 0; row < 5; row++ {
			for col := 0; col < 5; col++ {
				for split := 0; split < 2; split++ {
					if g.Value(litSplit(row, col, split)) {
						fmt.Printf("split %d at (%d,%d)\n", split, row, col)
					}
				}
			}
		}
	} else {
		fmt.Println("unsat!!")
	}
}
