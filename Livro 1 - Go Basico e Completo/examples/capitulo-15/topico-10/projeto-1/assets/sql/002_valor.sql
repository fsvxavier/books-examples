ALTER TABLE pedidos ADD CONSTRAINT valor_positivo CHECK (valor > 0);
