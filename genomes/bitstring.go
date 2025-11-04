package genomes

import "math/rand"

type BitString []bool

func NewBitString(length int) BitString {
    bs := make(BitString, length)
    for i := range bs {
        bs[i] = rand.Float64() < 0.5
    }
    return bs
}

func SinglePointCrossover(p1, p2 BitString) (BitString, BitString) {
    point := rand.Intn(len(p1))
    
    c1 := make(BitString, len(p1))
    c2 := make(BitString, len(p2))
    
    copy(c1[:point], p1[:point])
    copy(c1[point:], p2[point:])
    
    copy(c2[:point], p2[:point])
    copy(c2[point:], p1[point:])
    
    return c1, c2
}

func Mutate(bs BitString) BitString {
    mutated := make(BitString, len(bs))
    copy(mutated, bs)
    
    bit := rand.Intn(len(mutated))
    mutated[bit] = !mutated[bit]
    
    return mutated
}
