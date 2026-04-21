package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name        string `yaml:"name"`
	LLMProvider string `yaml:"llm_provider"`
	Model      string `yaml:"model"`
	Memory     string `yaml:"memory"`
	Workspace  string `yaml:"workspace"`
}

type ProfileStorage struct {
	profiles map[string]*Profile
	mu       sync.RWMutex
}

func NewProfile(name string) *Profile {
	return &Profile{
		Name:        name,
		LLMProvider: "openai",
		Model:       "gpt-4o",
		Memory:      "sqlite",
		Workspace:   "default",
	}
}

func NewProfileStorage() *ProfileStorage {
	return &ProfileStorage{profiles: make(map[string]*Profile)}
}

func (ps *ProfileStorage) Save(name string, profile *Profile) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.profiles[name] = profile
	return nil
}

func (ps *ProfileStorage) Load(name string) (*Profile, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.profiles[name]
	return p, ok
}

func (ps *ProfileStorage) List() []*Profile {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	result := make([]*Profile, 0, len(ps.profiles))
	for _, p := range ps.profiles {
		result = append(result, p)
	}
	return result
}

func (ps *ProfileStorage) Delete(name string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.profiles, name)
}

type ProfileSwitcher struct {
	current  *Profile
	profiles map[string]*Profile
	mu       sync.RWMutex
}

func NewProfileSwitcher() *ProfileSwitcher {
	return &ProfileSwitcher{profiles: make(map[string]*Profile)}
}

func (sw *ProfileSwitcher) AddProfile(p *Profile) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.profiles[p.Name] = p
}

func (sw *ProfileSwitcher) SwitchTo(name string) *Profile {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if p, ok := sw.profiles[name]; ok {
		sw.current = p
		return p
	}
	return nil
}

func (sw *ProfileSwitcher) Current() *Profile {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.current
}

type ProfileFileStorage struct {
	filePath string
	profiles map[string]*Profile
}

func NewProfileFileStorage(path string) *ProfileFileStorage {
	return &ProfileFileStorage{
		filePath: path,
		profiles: make(map[string]*Profile),
	}
}

func (fs *ProfileFileStorage) SaveToFile(profile *Profile) error {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filePath, data, 0644)
}

func (fs *ProfileFileStorage) LoadFromFile() (*Profile, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var profile Profile
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (fs *ProfileFileStorage) SaveAll(profiles []*Profile) error {
	data, err := yaml.Marshal(profiles)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filePath, data, 0644)
}

func (fs *ProfileFileStorage) LoadAll() ([]*Profile, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var profiles []*Profile
	err = yaml.Unmarshal(data, &profiles)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}