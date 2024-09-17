package images

const (
	featuresPrompt = `Provide me a list of 10 brief features for this image. "Should be in this brief style: A table with cupcakes and cake
A pink cake with a pig face on top
A white cake with a pig on it" The response should be a JSON array of strings. There should be nothing before and after the opening and closing array brackets.`
)

type AiImageModel string

const (
	StableImageCore AiImageModel = "core"
)
