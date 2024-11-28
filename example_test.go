package huffman

import "fmt"

func ExampleLabel() {
	// alphabet is the set of symbols to use for the labels.
	// We'll use a short alphabet for this example.
	const alphabet = "asdf"
	// We have some number of items, each with a frequency.
	// Frequency can be thought of as the priority of the item.
	// Higher priority items get shorter labels.
	items := []struct {
		value string
		freq  int
	}{
		{"parameter", 52},
		{"values", 52},
		{"variable", 55},
		{"argument", 56},
		{"slice", 59},
		{"types", 70},
		{"expression", 88},
		{"function", 158},
		{"value", 169},
		{"type", 530},
	}

	// Convert the items to a list of frequencies.
	freqs := make([]int, len(items))
	for i, item := range items {
		freqs[i] = item.freq
	}

	// Generate the labels.
	// Highest frequency items like "value" and "type"
	// will get the shortest labels.
	labels := Label(len(alphabet), freqs)

	// Print the labels.
	for i, labelIndexes := range labels {
		item := items[i]

		// labels[i] refers to indexes in alphabet.
		// Join them to form the full label.
		label := make([]byte, len(labelIndexes))
		for j, idx := range labelIndexes {
			label[j] = alphabet[idx]
		}

		fmt.Printf("%s: %s\n", item.value, label)
	}

	// Output:
	// parameter: sa
	// values: ss
	// variable: sd
	// argument: sf
	// slice: da
	// types: ds
	// expression: dd
	// function: df
	// value: a
	// type: f
}
