package images

const (
	featuresPrompt = `Provide me a list of 10 brief features for this image. "Should be in this brief style: A table with cupcakes and cake
A pink cake with a pig face on top
A white cake with a pig on it" The response should be a JSON array of strings. There should be nothing before and after the opening and closing array brackets.`

	bestImagePrompt = `I am going to show you a list of images. Here is a description "%s". Please reply with the index (counting from 0) of th image which best matches the description. If there is no appropriate image, reply with -1. The response should be just the index number.`
)

type AiImageModel string

const (
	StableImageCore AiImageModel = "core"
)
