package integrations

import (
	"context"
	"fmt"
	"strings"
)

var careTipsCatalog = []CareTip{
	{ID: "food-water", Category: "Alimentación", Emoji: "🍗", Title: "Agua fresca", Text: "Renueva el agua todos los días y mantén más de un punto de hidratación si tu gato no bebe mucho.", Source: "Mi Michi"},
	{ID: "food-portions", Category: "Alimentación", Emoji: "🍗", Title: "Porciones", Text: "Divide la comida diaria en 2 o 3 porciones para evitar atracones y controlar mejor el peso.", Source: "Mi Michi"},
	{ID: "food-transition", Category: "Alimentación", Emoji: "🍗", Title: "Cambio de alimento", Text: "Cuando cambies alimento, mezcla el nuevo con el anterior durante 7 a 10 días para reducir malestares digestivos.", Source: "Mi Michi"},
	{ID: "food-treats", Category: "Alimentación", Emoji: "🍗", Title: "Premios", Text: "Usa premios pequeños y descuéntalos de la ración diaria si estás trabajando control de peso.", Source: "Mi Michi"},
	{ID: "hygiene-litter", Category: "Higiene", Emoji: "🛁", Title: "Arenero limpio", Text: "Retira desechos del arenero todos los días y lava la bandeja de forma regular para evitar rechazo o estrés.", Source: "Mi Michi"},
	{ID: "hygiene-brushing", Category: "Higiene", Emoji: "🛁", Title: "Cepillado", Text: "Cepilla a tu gato varias veces por semana; en gatos de pelo largo conviene hacerlo con más frecuencia.", Source: "Mi Michi"},
	{ID: "hygiene-nails", Category: "Higiene", Emoji: "🛁", Title: "Uñas", Text: "Revisa las uñas cada 2 a 3 semanas y corta solo la punta transparente con herramientas adecuadas.", Source: "Mi Michi"},
	{ID: "play-hunt", Category: "Juego", Emoji: "🎾", Title: "Cazar jugando", Text: "Prioriza juguetes que simulan presas, con sesiones cortas de persecución y cierre con recompensa.", Source: "Mi Michi"},
	{ID: "play-daily", Category: "Juego", Emoji: "🎾", Title: "Rutina diaria", Text: "Jugar 10 a 15 minutos al día ayuda a reducir estrés, aburrimiento y exceso de energía nocturna.", Source: "Mi Michi"},
	{ID: "play-rotation", Category: "Juego", Emoji: "🎾", Title: "Rotación", Text: "Guarda algunos juguetes y rótalos cada pocos días para que no pierdan novedad.", Source: "Mi Michi"},
	{ID: "health-vet", Category: "Salud", Emoji: "❤️", Title: "Chequeo anual", Text: "Agenda al menos un control veterinario al año, incluso si tu gato parece sano.", Source: "Mi Michi"},
	{ID: "health-weight", Category: "Salud", Emoji: "❤️", Title: "Peso", Text: "Registra el peso una vez al mes; cambios rápidos pueden ser una señal temprana de problemas de salud.", Source: "Mi Michi"},
	{ID: "health-sterilization", Category: "Salud", Emoji: "❤️", Title: "Esterilización", Text: "La esterilización puede reducir riesgos de enfermedades y conductas asociadas al celo o marcaje.", Source: "Mi Michi"},
	{ID: "behavior-hiding", Category: "Comportamiento", Emoji: "🧠", Title: "Se esconde", Text: "Si tu gato se esconde más de lo habitual, revisa cambios en casa, apetito, dolor o signos de enfermedad.", Source: "Mi Michi"},
	{ID: "behavior-purring", Category: "Comportamiento", Emoji: "🧠", Title: "Ronroneo", Text: "El ronroneo puede aparecer por bienestar, estrés o dolor; interpreta siempre el contexto completo.", Source: "Mi Michi"},
	{ID: "behavior-scratching", Category: "Comportamiento", Emoji: "🧠", Title: "Rascado", Text: "Ubica rascadores firmes cerca de zonas de descanso o muebles que intenta marcar.", Source: "Mi Michi"},
	{ID: "home-toxic-plants", Category: "Hogar", Emoji: "🏠", Title: "Plantas", Text: "Evita plantas tóxicas como lirios, pothos y aloe vera, o mantenlas fuera de su alcance real.", Source: "Mi Michi"},
	{ID: "home-vertical", Category: "Hogar", Emoji: "🏠", Title: "Altura", Text: "Ofrece espacios altos y estables para que pueda observar, descansar y sentirse seguro.", Source: "Mi Michi"},
	{ID: "home-safe-zone", Category: "Hogar", Emoji: "🏠", Title: "Zona segura", Text: "Ten una zona tranquila con agua, cama y arenero disponible cuando haya visitas, ruido o mudanzas.", Source: "Mi Michi"},
	{ID: "breed-fit", Category: "Razas", Emoji: "📚", Title: "Antes de elegir raza", Text: "Considera energía, cepillado, salud y rutina familiar antes de elegir o comparar una raza.", Source: "Mi Michi"},
	{ID: "breed-mixed", Category: "Razas", Emoji: "📚", Title: "Mestizos", Text: "Los gatos mestizos también necesitan seguimiento por edad, peso, pelo y conducta; no dependen de una raza para recibir buen cuidado.", Source: "Mi Michi"},
}

func SearchCareTips(ctx context.Context, query, category string, limit int) ([]CareTip, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	search := searchable(query)
	category = strings.TrimSpace(category)
	results := make([]CareTip, 0, limit)
	seen := map[string]bool{}

	for _, tip := range careTipsCatalog {
		if !matchesCategory(tip.Category, category) || !matchesTipSearch(tip, search) {
			continue
		}
		appendTip(&results, seen, tip, limit)
	}

	if len(results) < limit && shouldIncludeBreedTips(category, search) {
		breedTips, err := buildBreedCareTips(ctx, search, category, limit-len(results))
		if err == nil {
			for _, tip := range breedTips {
				appendTip(&results, seen, tip, limit)
			}
		}
	}

	return results, nil
}

func buildBreedCareTips(ctx context.Context, search, category string, limit int) ([]CareTip, error) {
	if limit <= 0 {
		return []CareTip{}, nil
	}
	breeds, err := ListCatBreeds(ctx)
	if err != nil {
		return nil, err
	}

	tips := make([]CareTip, 0, limit)
	for _, breed := range breeds {
		name := spanishBreedName(breed)
		breedSearch := searchable(name + " " + breed.Name + " " + breed.Origin + " " + breed.Temperament)
		if search != "" && !strings.Contains(breedSearch, search) {
			continue
		}

		tip := CareTip{
			ID:       "breed-" + breed.ID,
			Category: "Razas",
			Emoji:    "📚",
			Title:    "Perfil de " + name,
			Text:     breedCareText(name, breed),
			Source:   "TheCatAPI",
		}
		if !matchesCategory(tip.Category, category) && !matchesCategory("Razas", category) {
			continue
		}
		tips = append(tips, tip)
		if len(tips) >= limit {
			break
		}
	}
	return tips, nil
}

func breedCareText(name string, breed CatBreed) string {
	parts := []string{}
	if breed.EnergyLevel >= 4 {
		parts = append(parts, "planifica juego diario intenso")
	} else if breed.EnergyLevel > 0 {
		parts = append(parts, "mantén sesiones de juego breves y constantes")
	}
	if breed.Grooming >= 4 {
		parts = append(parts, "refuerza el cepillado semanal")
	}
	if breed.HealthIssues >= 3 {
		parts = append(parts, "mantén controles veterinarios preventivos")
	}
	if breed.Hypoallergenic == 1 {
		parts = append(parts, "puede ser una opción a evaluar en hogares con sensibilidad a alergias")
	}
	if len(parts) == 0 {
		parts = append(parts, "ajusta alimentación, enriquecimiento y revisiones según edad y estilo de vida")
	}

	detail := strings.Join(parts, "; ")
	if breed.LifeSpan != "" {
		return fmt.Sprintf("Según los datos de raza de TheCatAPI, para %s conviene %s. Su esperanza de vida referencial es de %s años.", name, detail, breed.LifeSpan)
	}
	return fmt.Sprintf("Según los datos de raza de TheCatAPI, para %s conviene %s.", name, detail)
}

func appendTip(results *[]CareTip, seen map[string]bool, tip CareTip, limit int) {
	if len(*results) >= limit || seen[tip.ID] {
		return
	}
	seen[tip.ID] = true
	*results = append(*results, tip)
}

func matchesTipSearch(tip CareTip, search string) bool {
	if search == "" {
		return true
	}
	value := searchable(tip.Category + " " + tip.Title + " " + tip.Text)
	return strings.Contains(value, search)
}

func matchesCategory(value, category string) bool {
	category = strings.TrimSpace(category)
	if category == "" || strings.EqualFold(category, "Todos") {
		return true
	}
	return searchable(value) == searchable(category)
}

func shouldIncludeBreedTips(category, search string) bool {
	if search != "" {
		return true
	}
	return matchesCategory("Razas", category)
}

func searchable(value string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
		"é", "e", "è", "e", "ë", "e", "ê", "e",
		"í", "i", "ì", "i", "ï", "i", "î", "i",
		"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
		"ú", "u", "ù", "u", "ü", "u", "û", "u",
		"ñ", "n",
	)
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func spanishBreedName(breed CatBreed) string {
	if name, ok := breedNamesByID[breed.ID]; ok {
		return name
	}
	if name, ok := breedNamesByEnglish[strings.ToLower(breed.Name)]; ok {
		return name
	}
	return breed.Name
}

var breedNamesByID = map[string]string{
	"abys": "Abisinio", "aege": "Egeo", "abob": "Bobtail americano", "acur": "Curl americano",
	"asho": "Americano de pelo corto", "awir": "Americano de pelo duro", "amau": "Mau árabe",
	"bali": "Balinés", "beng": "Bengalí", "birm": "Birmano", "bomb": "Bombay",
	"bslo": "Británico de pelo largo", "bsho": "Británico de pelo corto", "bure": "Burmés",
	"buri": "Burmilla", "csho": "Colorpoint de pelo corto", "crex": "Cornish Rex",
	"cymr": "Cymric", "cypr": "Chipriota", "drex": "Devon Rex", "emau": "Mau egipcio",
	"ebur": "Burmés europeo", "esho": "Exótico de pelo corto", "hima": "Himalayo",
	"jbob": "Bobtail japonés", "java": "Javanés", "kuri": "Bobtail kuriliano",
	"mcoo": "Maine Coon", "mala": "Malayo", "norw": "Bosque de Noruega",
	"pers": "Persa", "rblu": "Azul ruso", "sfol": "Escocés plegado", "siam": "Siamés",
	"sibe": "Siberiano", "soma": "Somalí", "tonk": "Tonkinés", "tang": "Angora turco",
	"tvan": "Van turco",
}

var breedNamesByEnglish = map[string]string{
	"abyssinian": "Abisinio", "american bobtail": "Bobtail americano",
	"american curl": "Curl americano", "american shorthair": "Americano de pelo corto",
	"balinese": "Balinés", "bengal": "Bengalí", "birman": "Birmano",
	"british shorthair": "Británico de pelo corto", "burmese": "Burmés",
	"egyptian mau": "Mau egipcio", "exotic shorthair": "Exótico de pelo corto",
	"himalayan": "Himalayo", "japanese bobtail": "Bobtail japonés",
	"norwegian forest cat": "Bosque de Noruega", "persian": "Persa",
	"russian blue": "Azul ruso", "scottish fold": "Escocés plegado",
	"siamese": "Siamés", "siberian": "Siberiano", "somali": "Somalí",
	"turkish angora": "Angora turco", "turkish van": "Van turco",
}
