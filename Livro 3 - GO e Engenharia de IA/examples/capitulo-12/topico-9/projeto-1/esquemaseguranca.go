package longo

// EsquemaSeguranca liga o isolamento por usuário DENTRO do banco (row-level
// security) e cria o papel com que a aplicação se conecta. As políticas leem
// app.usuario, que a Loja define no começo de cada transação.
//
// Atenção: o dono da tabela não é afetado pela política (a menos que se use
// FORCE ROW LEVEL SECURITY). Por isso a aplicação NÃO deve conectar como dono.
const EsquemaSeguranca = `
ALTER TABLE memorias  ENABLE ROW LEVEL SECURITY;
ALTER TABLE entidades ENABLE ROW LEVEL SECURITY;
ALTER TABLE relacoes  ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS por_usuario ON memorias;
DROP POLICY IF EXISTS por_usuario ON entidades;
DROP POLICY IF EXISTS por_usuario ON relacoes;
CREATE POLICY por_usuario ON memorias  USING (usuario = current_setting('app.usuario', true)) WITH CHECK (usuario = current_setting('app.usuario', true));
CREATE POLICY por_usuario ON entidades USING (usuario = current_setting('app.usuario', true)) WITH CHECK (usuario = current_setting('app.usuario', true));
CREATE POLICY por_usuario ON relacoes  USING (usuario = current_setting('app.usuario', true)) WITH CHECK (usuario = current_setting('app.usuario', true));

DO $$
BEGIN
	IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'agente_app') THEN
		CREATE ROLE agente_app LOGIN PASSWORD 'app';
	END IF;
END $$;
GRANT SELECT, INSERT, UPDATE, DELETE ON memorias, entidades, relacoes TO agente_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO agente_app;
`
