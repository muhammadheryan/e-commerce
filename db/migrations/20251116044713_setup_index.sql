-- migrate:up
CREATE INDEX idx_users_email ON user(email);
CREATE INDEX idx_users_phone ON user(phone);

CREATE INDEX idx_shop_name ON shop(name);

CREATE INDEX idx_warehouse_shop ON warehouse(shop_id, status);
CREATE INDEX idx_warehouse_status ON warehouse(status);

CREATE INDEX idx_product_shop ON product(shop_id);
CREATE INDEX idx_product_name ON product(name);

CREATE INDEX idx_ws_product ON warehouse_stock(product_id);

CREATE INDEX idx_orders_user ON orders(user_id);

CREATE INDEX idx_order_item_order ON order_item(order_id);

CREATE INDEX idx_reservation_order ON stock_reservation(order_id);
CREATE INDEX idx_reservation_product ON stock_reservation(product_id);
CREATE INDEX idx_reservation_expire ON stock_reservation(expires_at);


-- migrate:down

