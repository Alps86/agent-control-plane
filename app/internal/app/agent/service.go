package agent

import (
	"context"
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
	portagent "agentcontrolplane/app/internal/port/agent"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	"github.com/google/uuid"
)

// NewService bindet Identität, Organisationsspeicher und Agentenspeicher.
func NewService(store portagent.Store, identity portorganisation.Identity, organizations portorganisation.Store) *Service {
	return &Service{store: store, identity: identity, organizations: organizations,
		catalog: domainagent.NewCatalog(), probes: map[domainagent.ExecutionKind]portagent.ReadinessProbe{
			domainagent.Eino: unconfiguredProbe{code: "model_unconfigured", reason: "Modellverbindung fehlt"},
		}}
}

// RegisterTemplate ergänzt eine Adaptervorlage ausschließlich beim Bootstrap.
func (s *Service) RegisterTemplate(template domainagent.Template) bool {
	if s == nil || s.catalog == nil {
		return false
	}

	return s.catalog.Register(template)
}

// RegisterProbe ergänzt einen adaptereigenen, seiteneffektfreien Bereitschaftstest.
func (s *Service) RegisterProbe(kind domainagent.ExecutionKind, probe portagent.ReadinessProbe) {
	if s == nil || probe == nil || kind == "" {
		return
	}

	s.probes[kind] = probe
}

// Create speichert genau einen Agenten aus einer kuratierten Vorlage.
func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (Profile, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return Profile{}, err
	}

	agent, err := s.prepare(organizationID, input)
	if err != nil {
		return Profile{}, err
	}

	if err := s.store.CreateAgent(ctx, operatorID, agent); err != nil {
		return Profile{}, err
	}

	return s.profile(ctx, agent), nil
}

// List liefert nur Agenten der eigenen Organisation.
func (s *Service) List(ctx context.Context, organizationID string) ([]Profile, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	agents, err := s.store.ListAgents(ctx, organizationID, operatorID)
	if err != nil {
		return nil, err
	}

	profiles := make([]Profile, 0, len(agents))
	for _, agent := range agents {
		profiles = append(profiles, s.profile(ctx, agent))
	}

	return profiles, nil
}

// Get liest einen Agenten nur innerhalb seiner Organisation.
func (s *Service) Get(ctx context.Context, organizationID, agentID string) (Profile, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return Profile{}, err
	}

	agent, err := s.store.FindAgent(ctx, organizationID, agentID, operatorID)
	if err != nil {
		return Profile{}, err
	}

	return s.profile(ctx, agent), nil
}

// Templates zeigt die kuratierte Auswahl nur in einer zugänglichen Organisation.
func (s *Service) Templates(ctx context.Context, organizationID string) ([]domainagent.Template, error) {
	if _, err := s.authorize(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.catalog.Templates(), nil
}

// TeamTemplates zeigt auswählbare Einzelprofile in Teamvorlagen.
func (s *Service) TeamTemplates(ctx context.Context, organizationID string) ([]domainagent.TeamTemplate, error) {
	if _, err := s.authorize(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.catalog.TeamTemplates(), nil
}

// CanStart liefert einen überprüfbaren Startentscheid ohne Laufanlage.
func (s *Service) CanStart(ctx context.Context, organizationID, agentID string) (domainagent.Readiness, error) {
	profile, err := s.Get(ctx, organizationID, agentID)
	if err != nil {
		return domainagent.Readiness{}, err
	}

	return profile.Readiness, nil
}

// Organization liest den Namen nur aus der serverseitig zugeordneten Organisation.
func (s *Service) Organization(ctx context.Context, organizationID string) (domainorganisation.Organization, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return domainorganisation.Organization{}, err
	}

	organization, err := s.organizations.Get(ctx, organizationID, operatorID)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return domainorganisation.Organization{}, ErrNotFound
	}

	return organization, err
}

func (s *Service) authorize(ctx context.Context, organizationID string) (string, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return "", err
	}

	if _, err := s.organizations.Get(ctx, organizationID, operatorID); err != nil {
		if errors.Is(err, portorganisation.ErrNotFound) {
			return "", ErrNotFound
		}

		return "", err
	}

	return operatorID, nil
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.store == nil || s.identity == nil || s.organizations == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}

func (s *Service) prepare(organizationID string, input CreateInput) (domainagent.Agent, error) {
	template, found := s.catalog.Find(input.TemplateID)
	if !found {
		return domainagent.Agent{}, ErrTemplateNotFound
	}

	if input.ExecutionKind != "" && input.ExecutionKind != template.Kind {
		return domainagent.Agent{}, ErrExecutionKind
	}

	agent, err := domainagent.NewAgent(uuid.NewString(), organizationID, input.Name, template)
	if err != nil {
		return domainagent.Agent{}, err
	}

	return agent, s.checkCapabilities(agent, input.AdditionalCapabilities)
}

func (s *Service) checkCapabilities(agent domainagent.Agent, additional []string) error {
	for _, capability := range additional {
		if !agent.AllowsCapability(capability) {
			return ErrCapabilityDenied
		}
	}

	return nil
}

func (s *Service) profile(ctx context.Context, agent domainagent.Agent) Profile {
	probe := s.probes[agent.ExecutionKind]
	if probe == nil {
		return Profile{Agent: agent, Readiness: domainagent.Readiness{
			Code: "profile_unverified", Reason: "Ausführungsprofil nicht nachgewiesen"}}
	}

	return Profile{Agent: agent, Readiness: probe.Status(ctx, agent)}
}

func (p unconfiguredProbe) Status(_ context.Context, _ domainagent.Agent) domainagent.Readiness {
	return domainagent.Readiness{Code: p.code, Reason: p.reason}
}
