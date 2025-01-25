package internal

// LoaderRegistry holds a map of extension => DataLoader
type LoaderRegistry struct {
	registry map[string]DataLoader
}

// GetLoader returns the DataLoader for a given file extension, if found.
func (lr *LoaderRegistry) GetLoader(ext string) (DataLoader, bool) {
	loader, ok := lr.registry[ext]
	return loader, ok
}

// RegisterLoader lets you add or overwrite a DataLoader for a specific extension.
func (lr *LoaderRegistry) RegisterLoader(ext string, loader DataLoader) {
	lr.registry[ext] = loader
}

// NewLoaderRegistry initializes a default registry with a CSV loader.
// You could easily extend this with JSONLoader, RepoLoader, etc.
func NewLoaderRegistry() *LoaderRegistry {
	return &LoaderRegistry{
		registry: map[string]DataLoader{
			".csv":  CSVLoader{},
			".json": JSONLoader{},
			// ".sql": RepoLoader{ DB: ... },
		},
	}
}

// Load all files in directory

// DataLoaderService orchestrates reading multiple files from a directory
// using the loaders from the LoaderRegistry.
type DataLoaderService struct {
	registry *LoaderRegistry
}

// NewDataLoaderService constructs the service with a LoaderRegistry
func NewDataLoaderService(lr *LoaderRegistry) *DataLoaderService {
	return &DataLoaderService{registry: lr}
}
