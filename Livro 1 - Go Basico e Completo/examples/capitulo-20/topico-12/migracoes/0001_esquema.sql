CREATE TABLE produtos (
    id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sku     TEXT   NOT NULL UNIQUE,
    nome    TEXT   NOT NULL,
    preco   BIGINT NOT NULL CHECK (preco >= 0),
    estoque INT    NOT NULL CHECK (estoque >= 0)
);

CREATE TABLE pedidos (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cliente   TEXT        NOT NULL,
    total     BIGINT      NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE itens_pedido (
    pedido_id  BIGINT NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    produto_id BIGINT NOT NULL REFERENCES produtos(id),
    qtd        INT    NOT NULL CHECK (qtd > 0),
    preco      BIGINT NOT NULL,
    PRIMARY KEY (pedido_id, produto_id)
);

CREATE INDEX pedidos_criado_em_idx ON pedidos (criado_em DESC, id DESC);
