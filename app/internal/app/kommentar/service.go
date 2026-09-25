package kommentar

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	"agentcontrolplane/app/internal/domain/rechte"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	"github.com/google/uuid"
)

// NewService bindet Speicher und die zwei serverseitigen Identitätsquellen.
func NewService(store portkommentar.Store, operatorIdentity, agentIdentity portorganisation.Identity, resolver ...portkommentar.ArtifactResolver) *Service {
	service := &Service{store: store, operatorIdentity: operatorIdentity, agentIdentity: agentIdentity}
	if len(resolver) == 1 {
		service.artifactResolver = resolver[0]
	}

	return service
}

// Create fügt als lokaler Betreiber einen Kommentar hinzu.
func (s *Service) Create(ctx context.Context, orgID, taskID string, input CreateInput) (domainkommentar.Comment, error) {
	actor, err := s.trustedActor(rechte.Operator)
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	return s.create(ctx, actor, orgID, taskID, input)
}

// CreateForAgent prüft die benannte Fachaktion und die vertrauenswürdige Agentenidentität.
func (s *Service) CreateForAgent(ctx context.Context, orgID, taskID, actionID string, input CreateInput) (domainkommentar.Comment, error) {
	if actionID != ActionTaskComment {
		return domainkommentar.Comment{}, ErrActionDenied
	}

	actor, err := s.trustedActor(rechte.Agent)
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	return s.create(ctx, actor, orgID, taskID, input)
}

func (s *Service) create(ctx context.Context, actor rechte.Actor, orgID, taskID string, input CreateInput) (domainkommentar.Comment, error) {
	content, err := (domainkommentar.Comment{Content: input.Content}).ValidatedContent()
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	reference, err := input.Reference.Validate()
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	reference, err = s.reference(ctx, orgID, taskID, reference)
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	comment := domainkommentar.Comment{ID: uuid.NewString(), OrganizationID: orgID, TaskID: taskID,
		Content: content, SourceKind: string(actor.Kind), SourceID: actor.ID, Reference: reference}
	created, err := s.store.CreateComment(ctx, actor, comment)
	return created, s.audit(ctx, actor, orgID, taskID, "create", err)
}

func (s *Service) reference(ctx context.Context, orgID, taskID string, reference *domainkommentar.Reference) (*domainkommentar.Reference, error) {
	if reference == nil {
		return nil, nil
	}

	if reference.OrganizationID != orgID {
		return nil, ErrInvalidReference
	}

	if reference.Type == domainkommentar.ReferenceTask {
		return reference, nil
	}

	return s.artifactReference(ctx, orgID, taskID, reference)
}

func (s *Service) artifactReference(ctx context.Context, orgID, taskID string, reference *domainkommentar.Reference) (*domainkommentar.Reference, error) {
	if s.artifactResolver == nil {
		return nil, ErrInvalidReference
	}

	meta, err := s.artifactResolver.ResolveArtifact(ctx, orgID, taskID, reference.ID)
	if err != nil {
		return nil, ErrInvalidReference
	}

	if !s.validArtifactMetadata(meta, orgID, taskID, reference.ID) {
		return nil, ErrInvalidReference
	}

	reference.DisplayName, reference.Link = meta.DisplayName, meta.Link
	return reference, nil
}

func (s *Service) validArtifactMetadata(meta portkommentar.ArtifactMetadata, orgID, taskID, artifactID string) bool {
	return meta.ID == artifactID && meta.OrganizationID == orgID && meta.TaskID == taskID &&
		strings.TrimSpace(meta.DisplayName) != "" && utf8.RuneCountInString(meta.DisplayName) <= 256 &&
		strings.IndexFunc(meta.DisplayName, unicode.IsControl) < 0 && internalLinkPattern.MatchString(meta.Link)
}

// List liest den Aufgabenthread als Betreiber.
func (s *Service) List(ctx context.Context, orgID, taskID string) ([]domainkommentar.Comment, error) {
	actor, err := s.trustedActor(rechte.Operator)
	if err != nil {
		return nil, err
	}

	return s.list(ctx, actor, orgID, taskID)
}

// ListForAgent liest nur im aktuellen Projektbereich des vertrauenswürdigen Agenten.
func (s *Service) ListForAgent(ctx context.Context, orgID, taskID string) ([]domainkommentar.Comment, error) {
	actor, err := s.trustedActor(rechte.Agent)
	if err != nil {
		return nil, err
	}

	return s.list(ctx, actor, orgID, taskID)
}

func (s *Service) list(ctx context.Context, actor rechte.Actor, orgID, taskID string) ([]domainkommentar.Comment, error) {
	comments, err := s.store.ListComments(ctx, actor, orgID, taskID)
	return comments, s.audit(ctx, actor, orgID, taskID, "read", err)
}

// Update ändert nur den Inhalt; Quelle und Erstellzeit bleiben erhalten.
func (s *Service) Update(ctx context.Context, orgID, taskID, commentID, content string) (domainkommentar.Comment, error) {
	actor, err := s.trustedActor(rechte.Operator)
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	return s.update(ctx, actor, orgID, taskID, commentID, content)
}

// UpdateForAgent akzeptiert nur die benannte Fachaktion.
func (s *Service) UpdateForAgent(ctx context.Context, orgID, taskID, commentID, actionID, content string) (domainkommentar.Comment, error) {
	if actionID != ActionTaskComment {
		return domainkommentar.Comment{}, ErrActionDenied
	}

	actor, err := s.trustedActor(rechte.Agent)
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	return s.update(ctx, actor, orgID, taskID, commentID, content)
}

func (s *Service) update(ctx context.Context, actor rechte.Actor, orgID, taskID, commentID, content string) (domainkommentar.Comment, error) {
	content, err := (domainkommentar.Comment{Content: content}).ValidatedContent()
	if err != nil {
		return domainkommentar.Comment{}, err
	}

	updated, err := s.store.UpdateComment(ctx, actor, orgID, taskID, commentID, content)
	return updated, s.audit(ctx, actor, orgID, taskID, "update", err)
}

func (s *Service) trustedActor(kind rechte.ActorKind) (rechte.Actor, error) {
	if s == nil || s.store == nil {
		return rechte.Actor{}, ErrAccessDenied
	}

	identity := s.operatorIdentity
	if kind == rechte.Agent {
		identity = s.agentIdentity
	}

	return s.actor(identity, kind)
}

func (s *Service) actor(identity portorganisation.Identity, kind rechte.ActorKind) (rechte.Actor, error) {
	if identity == nil {
		return rechte.Actor{}, ErrAccessDenied
	}

	actors := identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != kind {
		return rechte.Actor{}, ErrAccessDenied
	}

	return actors[0], nil
}

func (s *Service) audit(ctx context.Context, actor rechte.Actor, orgID, taskID, action string, err error) error {
	if !errors.Is(err, ErrNotFound) || actor.Kind != rechte.Agent {
		return err
	}

	if auditErr := s.store.AuditDeniedComment(ctx, actor, orgID, taskID, action); auditErr != nil {
		return auditErr
	}

	return err
}
