INSERT INTO veterinarians ("id", "name", "address", "customer_ids")
VALUES (gen_random_uuid(), 'Dr. Bob', '123 Main St', ARRAY ['78858649-cc0d-11ed-b9e7-0242ac1b0003'::uuid, '78858996-cc0d-11ed-b9e7-0242ac1b0003'::uuid]),
       (gen_random_uuid(), 'Dr. Sue', '456 Main St', ARRAY ['78858649-cc0d-11ed-b9e7-0242ac1b0003'::uuid, '78858bee-cc0d-11ed-b9e7-0242ac1b0003'::uuid]),
       (gen_random_uuid(), 'Dr. Joe', '789 Main St', ARRAY ['78858ed1-cc0d-11ed-b9e7-0242ac1b0003'::uuid, '78858bee-cc0d-11ed-b9e7-0242ac1b0003'::uuid, '78858996-cc0d-11ed-b9e7-0242ac1b0003'::uuid]);

