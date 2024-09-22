package campaign_helper

const (
	featuresFromDescriptionPrompt = `Provide me a list of 10 brief possible captions for the image generated from this prompt: "%s". The captions should be in this brief style: A table with cupcakes and cake
	A pink cake with a pig face on top
	A white cake with a pig on it" The response should be a JSON array of strings. There should be nothing before and after the opening and closing array brackets.`
)
