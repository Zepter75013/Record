// Convention standard des coffrets vinyle : les faces sont numérotées en
// continu à travers les disques (Disque 1 = Faces A/B, Disque 2 = C/D,
// Disque 3 = E/F, Disque 4 = G/H…). La position d'une piste ("A1", "C3"…)
// commence par cette lettre de face suivie du numéro de piste sur la face.

const FACE_REGEX = /^([A-Z])/i

export function letterToDisc(letter) {
  if (!letter) return null
  const index = letter.toUpperCase().charCodeAt(0) - 65 // A=0, B=1, C=2…
  if (index < 0 || index > 25) return null
  return Math.floor(index / 2) + 1
}

export function letterToSide(letter) {
  if (!letter) return null
  const index = letter.toUpperCase().charCodeAt(0) - 65
  if (index < 0 || index > 25) return null
  return index % 2 === 0 ? 'A' : 'B'
}

export function discSideToLetter(disc, side) {
  const index = (disc - 1) * 2 + (side === 'B' ? 1 : 0)
  return String.fromCharCode(65 + index)
}

export function parsePosition(position) {
  const match = FACE_REGEX.exec((position || '').trim())
  if (!match) return { letter: null, disc: null, side: null }
  const letter = match[1].toUpperCase()
  return { letter, disc: letterToDisc(letter), side: letterToSide(letter) }
}

// Regroupe une liste de pistes par disque puis par face (A/B), en
// conservant l'ordre reçu à l'intérieur de chaque face. Les pistes dont la
// position ne commence pas par une lettre reconnaissable (vide, purement
// numérique…) sont renvoyées à part dans `noFace`.
export function groupTracksByDiscSide(tracks) {
  const discsMap = new Map() // disc -> { A: [...], B: [...] }
  const noFace = []

  for (const track of tracks || []) {
    const { disc, side } = parsePosition(track.position)
    if (disc === null) {
      noFace.push(track)
      continue
    }
    if (!discsMap.has(disc)) discsMap.set(disc, { A: [], B: [] })
    discsMap.get(disc)[side].push(track)
  }

  const discs = [...discsMap.keys()]
    .sort((a, b) => a - b)
    .map((disc) => {
      const sidesObj = discsMap.get(disc)
      return {
        disc,
        sides: ['A', 'B'].map((side) => ({
          side,
          letter: discSideToLetter(disc, side),
          tracks: sidesObj[side]
        }))
      }
    })

  return { discs, noFace }
}
