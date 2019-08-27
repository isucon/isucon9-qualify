-- mysql> alter table items add root_category_id int unsigned NOT NULL after category_id;
-- Query OK, 0 rows affected (1.60 sec)
-- Records: 0  Duplicates: 0  Warnings: 0
-- mysql> update items i join categories c on i.category_id=c.id set i.root_category_id = c.parent_id;
-- Query OK, 50162 rows affected (0.44 sec)
-- Rows matched: 50162  Changed: 50162  Warnings: 0
alter table items add root_category_id int unsigned NOT NULL after category_id;
alter table items add index idx_root_category(root_category_id,created_at);
update items i join categories c on i.category_id=c.id set i.root_category_id = c.parent_id;
