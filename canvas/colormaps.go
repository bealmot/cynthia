package canvas

// Perceptually uniform color maps for scientific visualization.
// These are monotonically increasing in luminance, colorblind-friendly,
// and readable in grayscale. Sampled at 16 stops from the original
// matplotlib implementations.

// Viridis is a perceptually uniform colormap from dark purple through
// teal to yellow. The default choice for most scalar field visualizations.
var Viridis = Palette{
	Hex("#440154"), Hex("#481567"), Hex("#482677"), Hex("#453781"),
	Hex("#3F4788"), Hex("#38588C"), Hex("#31688E"), Hex("#2A788E"),
	Hex("#23888E"), Hex("#1F988B"), Hex("#22A884"), Hex("#35B779"),
	Hex("#54C568"), Hex("#7AD151"), Hex("#A5DB36"), Hex("#FDE725"),
}

// Magma is a perceptually uniform colormap from black through deep
// purple and hot pink to light yellow. High contrast, dramatic feel.
var Magma = Palette{
	Hex("#000004"), Hex("#0C0926"), Hex("#221150"), Hex("#3B0F70"),
	Hex("#57106E"), Hex("#721F81"), Hex("#8C2981"), Hex("#A8327D"),
	Hex("#C43C75"), Hex("#DE4968"), Hex("#F1605D"), Hex("#F8835A"),
	Hex("#FCAE6B"), Hex("#FDCD90"), Hex("#FBECB2"), Hex("#FCFDBF"),
}

// Inferno is a perceptually uniform colormap from black through deep
// blue, fiery orange to bright yellow. Intense, high-energy feel.
var Inferno = Palette{
	Hex("#000004"), Hex("#0D0829"), Hex("#1F0C48"), Hex("#380962"),
	Hex("#550F6D"), Hex("#72186B"), Hex("#8C2369"), Hex("#A62F5E"),
	Hex("#BC3F52"), Hex("#CF5246"), Hex("#E06838"), Hex("#ED802E"),
	Hex("#F59B24"), Hex("#F9B91B"), Hex("#F9D71C"), Hex("#FCFFA4"),
}

// Cividis is a perceptually uniform colormap optimized for color vision
// deficiency. Goes from dark blue through gray to yellow.
var Cividis = Palette{
	Hex("#002051"), Hex("#0A326A"), Hex("#1E3F72"), Hex("#324D6E"),
	Hex("#425B6C"), Hex("#51696C"), Hex("#60786D"), Hex("#708770"),
	Hex("#819674"), Hex("#93A678"), Hex("#A5B67A"), Hex("#B9C77B"),
	Hex("#CDD87A"), Hex("#E2E975"), Hex("#F4F96E"), Hex("#FDFD66"),
}

// Plasma is a perceptually uniform colormap from deep purple through
// magenta and orange to bright yellow.
var Plasma = Palette{
	Hex("#0D0887"), Hex("#2D0594"), Hex("#4903A0"), Hex("#6500A7"),
	Hex("#7E03A8"), Hex("#9600A4"), Hex("#AB149E"), Hex("#BE2F93"),
	Hex("#CF4482"), Hex("#DD5D6E"), Hex("#E97858"), Hex("#F29441"),
	Hex("#F7B02B"), Hex("#F9CB1D"), Hex("#F5E626"), Hex("#F0F921"),
}

// Coolwarm is a diverging colormap centered on white, ranging from
// cool blue (negative) through neutral to warm red (positive).
// Ideal for data with a meaningful midpoint (zero).
var Coolwarm = Palette{
	Hex("#3B4CC0"), Hex("#4F6BD0"), Hex("#6788DC"), Hex("#82A3E4"),
	Hex("#9DBBE9"), Hex("#B5CFEE"), Hex("#CBDDF0"), Hex("#E0E8EF"),
	Hex("#EDE4E1"), Hex("#F0D3C5"), Hex("#ECBDA5"), Hex("#E3A285"),
	Hex("#D68568"), Hex("#C5674E"), Hex("#B14938"), Hex("#B40426"),
}

// Turbo is a rainbow-like colormap with improved perceptual uniformity.
// From dark blue through cyan, green, yellow, red to dark red.
var Turbo = Palette{
	Hex("#30123B"), Hex("#4145AB"), Hex("#4675ED"), Hex("#39A0FC"),
	Hex("#1FC8D8"), Hex("#29E9AE"), Hex("#6DFD62"), Hex("#A4FC3C"),
	Hex("#CDE927"), Hex("#EDCB13"), Hex("#FCA60C"), Hex("#F67D11"),
	Hex("#E55214"), Hex("#C92D14"), Hex("#A31212"), Hex("#7A0403"),
}
