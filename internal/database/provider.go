package database

import "context"

func (s *service) GetProvider(ctx context.Context, id string) (*Provider, error) {
	var p Provider
	query := `SELECT id, provider_name, provider_url, verification_pending, version, verified_at, provider_type, created_at FROM providers WHERE id = ?`
	err := s.db.GetContext(ctx, &p, query, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *service) GetProviderByName(ctx context.Context, name string) (*Provider, error) {
	var p Provider
	query := `SELECT id, provider_name, provider_url, verification_pending, version, verified_at, provider_type, created_at FROM providers WHERE provider_name = ?`
	err := s.db.GetContext(ctx, &p, query, name)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *service) UpsertProvider(ctx context.Context, p *Provider) error {
	query := `
		INSERT INTO providers (id, provider_name, provider_url, verification_pending, version, verified_at, provider_type, created_at)
		VALUES (:id, :provider_name, :provider_url, :verification_pending, :version, :verified_at, :provider_type, :created_at)
		ON CONFLICT(provider_name) DO UPDATE SET
			provider_url = excluded.provider_url,
			verification_pending = excluded.verification_pending,
			version = excluded.version,
			verified_at = excluded.verified_at,
			provider_type = excluded.provider_type
	`
	_, err := s.db.NamedExecContext(ctx, query, p)
	return err
}

func (s *service) ListProviders(ctx context.Context, limit, offset int) ([]Provider, error) {
	var list []Provider
	query := `SELECT id, provider_name, provider_url, verification_pending, version, verified_at, provider_type, created_at FROM providers ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &list, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return list, nil
}
