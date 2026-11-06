-- Schema canónico SurrealDB para Velo
DEFINE TABLE pending_action SCHEMAFULL;

DEFINE FIELD kind ON pending_action TYPE string;
DEFINE FIELD payload ON pending_action TYPE object;
DEFINE FIELD status ON pending_action TYPE string ASSERT $value IN ['pending', 'approved', 'rejected', 'expired'];
DEFINE FIELD created_at ON pending_action TYPE datetime VALUE time::now() READONLY;
DEFINE FIELD expires_at ON pending_action TYPE datetime;
