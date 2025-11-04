package genomes

import "math/rand/v2"

type BitString []bool

func NewBitString(length int) BitString {
    bs := make(BitString, length)
    for i := range bs {
        bs[i] = rand.Float64() < 0.5
    }
    return bs
}

func SinglePointCrossover(p1, p2 BitString) (BitString, BitString) {
    point := rand.IntN(len(p1))

    c1 := make(BitString, len(p1))
    c2 := make(BitString, len(p2))
    
    copy(c1[:point], p1[:point])
    copy(c1[point:], p2[point:])
    
    copy(c2[:point], p2[:point])
    copy(c2[point:], p1[point:])
    
    return c1, c2
}

func Mutate(bs BitString) BitString {
    bit := rand.IntN(len(bs))
    bs[bit] = !bs[bit]
    
    return bs
}
