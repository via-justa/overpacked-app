-- Test data for the dev environment.
-- Applied by the `test-data` compose service after migrations/seeds have run.
-- Idempotent: uses fixed UUIDs with ON CONFLICT DO NOTHING so it can run repeatedly.

-- Persons: five people with varied gender, age, body weight and conditioning level.
INSERT INTO persons (id, name, gender, birthdate, body_weight_grams, conditioning_level)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'Alex Carter',   'male',   '1991-04-12', 80000, 'athletic'),
    ('22222222-2222-2222-2222-222222222222', 'Maya Lindgren', 'female', '1997-09-30', 62000, 'average'),
    ('33333333-3333-3333-3333-333333333333', 'Tom Becker',    'male',   '1974-01-05', 95000, 'sedentary'),
    ('44444444-4444-4444-4444-444444444444', 'Sofia Reyes',   'female', '2002-06-21', 58000, 'military'),
    ('55555555-5555-5555-5555-555555555555', 'Jordan Lee',    'other',  '1985-11-17', 70000, 'athletic')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Item categories (item_types) beyond the 5 system profiles, with type fields.
-- These give the inventory below meaningful, category-specific properties.
-- ---------------------------------------------------------------------------
INSERT INTO item_types (id, name, base_profile, is_system) VALUES
    ('cookware',   'Cookware',   NULL, FALSE),
    ('water',      'Water',      NULL, FALSE),
    ('pack',       'Pack',       NULL, FALSE),
    ('footwear',   'Footwear',   NULL, FALSE),
    ('toiletries', 'Toiletries', NULL, FALSE),
    ('first_aid',  'First Aid',  NULL, FALSE),
    ('bike',       'Bike',       NULL, FALSE),
    ('camera',     'Camera',     NULL, FALSE),
    ('accessory',  'Accessory',  NULL, FALSE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO item_type_fields (item_type_id, field_key, field_label, data_type, is_required, enum_options_json, unit, sort_order) VALUES
    ('cookware', 'cook_type',   'Cook type',   'enum',    FALSE, '["stove","pot","pan","mug","bowl","utensil","grinder","other"]'::jsonb, NULL, 10),
    ('cookware', 'capacity_ml', 'Capacity',    'number',  FALSE, NULL, 'ml', 20),
    ('cookware', 'material',    'Material',    'string',  FALSE, NULL, NULL, 30),

    ('water', 'water_type',  'Water type',  'enum',    FALSE, '["filter","bottle","flask","reservoir","other"]'::jsonb, NULL, 10),
    ('water', 'capacity_ml', 'Capacity',    'number',  FALSE, NULL, 'ml', 20),
    ('water', 'insulated',   'Insulated',   'boolean', FALSE, NULL, NULL, 30),

    ('pack', 'pack_type',        'Pack type',        'enum',    FALSE, '["backpack","dry_bag","stuff_sack","pocket","rack","pannier"]'::jsonb, NULL, 10),
    ('pack', 'capacity_liters',  'Capacity',         'number',  FALSE, NULL, 'L', 20),
    ('pack', 'waterproof',       'Waterproof',       'boolean', FALSE, NULL, NULL, 30),

    ('footwear', 'footwear_type', 'Footwear type', 'enum',    FALSE, '["trail_runner","hiking_boot","sandal","casual","cycling"]'::jsonb, NULL, 10),
    ('footwear', 'waterproof',    'Waterproof',    'boolean', FALSE, NULL, NULL, 20),
    ('footwear', 'size',          'Size',          'string',  FALSE, NULL, NULL, 30),

    ('toiletries', 'toiletry_type', 'Toiletry type', 'enum',   FALSE, '["towel","brush","bag","tool","other"]'::jsonb, NULL, 10),
    ('toiletries', 'material',      'Material',      'string',  FALSE, NULL, NULL, 20),

    ('first_aid', 'kit_type', 'Kit type', 'enum',    FALSE, '["bandage","tape","tool","medication","kit"]'::jsonb, NULL, 10),
    ('first_aid', 'quantity', 'Quantity', 'integer', FALSE, NULL, NULL, 20),

    ('bike', 'bike_part', 'Bike part', 'enum', FALSE, '["tool","lock","sensor","rack","tube","accessory"]'::jsonb, NULL, 10),

    ('camera', 'camera_type', 'Camera type', 'enum', FALSE, '["camera","action_cam","tripod","case","accessory"]'::jsonb, NULL, 10),

    ('accessory', 'subtype', 'Subtype', 'string', FALSE, NULL, NULL, 10)
ON CONFLICT (item_type_id, field_key) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Manufacturers from the inventory that are not already seeded.
-- Existing manufacturers are reused by canonical name (case/spacing tolerant):
--   Altra (-> Altra Running), Schoeffel (-> Schöffel), plus Sea to Summit,
--   McKinley, Osprey, Icebreaker, Fjällräven, Garmin, Vaude, 3F UL Gear,
--   Buff, Deuter.
-- "Unbranded" is a placeholder for inventory rows with no manufacturer.
-- ---------------------------------------------------------------------------
INSERT INTO manufacturers (name, website) VALUES
    ('Unbranded',      NULL),
    ('Sawyer',         'https://sawyer.com'),
    ('Trekology',      'https://trekology.com'),
    ('Racktime',       'https://www.racktime.com'),
    ('Aeropress',      'https://aeropress.com'),
    ('Arkel',          'https://www.arkel-od.com'),
    ('Falke',          'https://www.falke.com'),
    ('Pedco',          NULL),
    ('Topeak',         'https://www.topeak.com'),
    ('RIAS',           NULL),
    ('Duco',           NULL),
    ('Canon',          'https://www.canon.com'),
    ('Tamrac',         'https://www.tamrac.com'),
    ('Alpina',         'https://www.alpina-sports.com'),
    ('Lejorain',       NULL),
    ('Amazon',         'https://www.amazon.com'),
    ('BTWIN',          'https://www.btwin.com'),
    ('Schwalbe',       'https://www.schwalbe.com'),
    ('Optimus',        'https://optimusstoves.com'),
    ('BiC',            'https://www.bic.com'),
    ('Sony',           'https://www.sony.com'),
    ('Rolling Square', 'https://rollingsquare.com'),
    ('Apple',          'https://www.apple.com'),
    ('Spigen',         'https://www.spigen.com'),
    ('Eider',          'https://www.eider.com'),
    ('Anker',          'https://www.anker.com'),
    ('Ortlieb',        'https://www.ortlieb.com'),
    ('Castelli',       'https://www.castelli-cycling.com'),
    ('Scott',          'https://www.scott-sports.com'),
    ('Hammerhead',     'https://www.hammerhead.io'),
    ('Northwave',      'https://www.northwave.com'),
    ('Abus',           'https://www.abus.com'),
    ('Nitecore',       'https://www.nitecore.com'),
    ('AfterShokz',     'https://shokz.com'),
    ('PackTowl',       'https://www.packtowl.com'),
    ('Darn Tough',     'https://darntough.com'),
    ('Porlex',         'https://porlex.co.jp'),
    ('DJI',            'https://www.dji.com'),
    ('PGYTECH',        'https://www.pgytech.com'),
    ('Homeet',         NULL),
    ('VIVOBAREFOOT',   'https://www.vivobarefoot.com'),
    ('Lixada',         'https://www.lixada.com'),
    ('Uniqlo',         'https://www.uniqlo.com'),
    ('Karl Loven',     NULL),
    ('Salt Of The Earth', 'https://saltoftheearthnatural.com'),
    ('Mykita',         'https://mykita.com'),
    ('BSN',            'https://www.bsnmedical.com'),
    ('Get Out Gear',   NULL),
    ('Cocoon',         'https://www.cocoon.at'),
    ('Decathlon',      'https://www.decathlon.com'),
    ('Beurer',         'https://www.beurer.com'),
    ('ProCase',        NULL),
    ('Philips',        'https://www.philips.com'),
    ('Arcade',         'https://arcadebelts.com'),
    ('Loop',           'https://www.loopearplugs.com'),
    ('Matador',        'https://matadorequipment.com'),
    ('Truetape',       NULL),
    ('Xero',           'https://www.xeroshoes.com'),
    ('Foldies',        'https://www.foldies.com'),
    ('Tooletries',     'https://tooletries.com.au'),
    ('Wise Pilgrim',   'https://wisepilgrim.com'),
    ('MCPAL',          NULL),
    ('Compeed',        'https://www.compeed.com'),
    ('Outdoor Edge',   'https://www.outdooredge.com'),
    ('Finpac',         NULL),
    ('Dynamics',       NULL)
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Inventory items.
-- Staged in a temp table so item rows and their labels stay in sync.
-- Columns: slug (stable id key), type_id (category), name, manufacturer
-- (canonical), weight_grams (NULL when unknown/0 due to >0 constraint),
-- price (0 -> NULL), source_url, description, attributes (category props),
-- label (existing label name).
-- ---------------------------------------------------------------------------
CREATE TEMP TABLE _seed_items (
    slug          TEXT,
    type_id       TEXT,
    name          TEXT,
    manufacturer  TEXT,
    weight_grams  NUMERIC,
    price         NUMERIC,
    source_url    TEXT,
    description   TEXT,
    attributes    JSONB,
    label         TEXT
);

-- Water system
INSERT INTO _seed_items VALUES
    ('water-filter',        'water', 'Water filter',         'Sawyer',        80,    35, 'https://sawyer.com/products/sawyer-micro-squeeze-water-filtration-system/', 'Micro Squeeze Water Filter', '{"water_type":"filter"}', 'water filter'),
    ('thermal-water-bottle','water', 'Thermal water bottle', 'Unbranded',     287.6,  0, '', '', '{"water_type":"bottle","insulated":true}', 'water bottle'),
    ('water-bottle-vaude',  'water', 'Water Bottle',         'Vaude',         75,     6, 'https://www.vaude.com', 'Bike Bottle Organic', '{"water_type":"bottle","capacity_ml":750}', 'water bottle'),
    ('water-bottle-700',    'water', 'Water bottle 700cc',   'MCPAL',         137,    0, '', '', '{"water_type":"bottle","capacity_ml":700}', 'water bottle'),
    ('soft-water-flask-500','water', 'Soft Water Flask 500cc','Decathlon',    35,    12, 'https://www.decathlon.de/p/trinkflasche-weich-trail-500-ml-blau/_/R-p-327568', 'Kalenji Soft Flask 500', '{"water_type":"flask","capacity_ml":500}', 'water bottle');

-- Cookware
INSERT INTO _seed_items VALUES
    ('coffee-maker',  'cookware', 'Coffee maker',  'Aeropress',     185, 27, 'https://aeropress.com/product/aeropress-coffee-maker/', 'Aeropress', '{"cook_type":"other"}', 'cooking pot'),
    ('mug-x',         'cookware', 'Mug',           'Sea to Summit', 45,  10, 'https://seatosummit.com', 'X-Mug', '{"cook_type":"mug","capacity_ml":250,"material":"silicone"}', 'mug'),
    ('lighter',       'cookware', 'Lighter',       'BiC',           21,   0, '', 'Mini Lighter', '{"cook_type":"other"}', 'lighter'),
    ('coffee-grinder','cookware', 'Coffee grinder','Porlex',        256, 64, '', 'Mini Metal Coffee Grinder', '{"cook_type":"grinder"}', 'cooking pot'),
    ('bowl-x',        'cookware', 'Bowl',          'Sea to Summit', 80,  17, 'https://seatosummit.com', 'X-Bowl', '{"cook_type":"bowl","material":"silicone"}', 'bowl'),
    ('stove',         'cookware', 'Stove',         'Lixada',        25,  22, 'https://www.lixada.com', 'Gas Stove', '{"cook_type":"stove"}', 'stove'),
    ('spork',         'cookware', 'Cutlery (Spork)','Outdoor Edge', 70,  40, 'https://www.outdooredge.com/products/chowpal', 'Chowpal — including bag (2g)', '{"cook_type":"utensil"}', 'spork'),
    ('cutlery-set',   'cookware', 'Cutlery Set',   'Sea to Summit', 42,  27, 'https://seatosummit.com', 'Alpha Cutlery Set', '{"cook_type":"utensil"}', 'eating utensils'),
    ('pot-x',         'cookware', 'Pot',           'Sea to Summit', 250, 39, 'https://seatosummit.com', 'X-Pot', '{"cook_type":"pot","capacity_ml":1400,"material":"silicone"}', 'cooking pot');

-- Consumables
INSERT INTO _seed_items VALUES
    ('solid-wet-wipes','consumable', 'Solid wet wipes', 'RIAS',             2,  0, '', 'Compressed Mini Hand Towel', '{"requires_water":true}', 'soap'),
    ('hand-sanitizer', 'consumable', 'Hand Sanitizer',  'Unbranded',       50,  0, '', '', '{}', 'soap'),
    ('toothpaste',     'consumable', 'Toothpaste',      'Unbranded',       29,  0, '', '', '{}', 'toothpaste'),
    ('toilet-paper',   'consumable', 'Toilet Paper',    'Unbranded',       46,  0, '', '', '{}', 'toilet paper'),
    ('sunscreen',      'consumable', 'Sunscreen',       'Unbranded',       75,  0, '', '', '{}', 'sunscreen'),
    ('gas',            'consumable', 'Gas',             'Optimus',        205,  0, '', 'Stove fuel canister', '{}', 'fuel canister'),
    ('deodorant',      'consumable', 'Deodorant',       'Salt Of The Earth',50, 9, 'https://saltoftheearthnatural.com', 'Natural Travel Size Crystal Deodorant', '{}', 'soap'),
    ('solid-body-soap','consumable', 'Solid Body Soap', 'Unbranded',       72,  0, '', 'Including waterproof bag', '{}', 'soap'),
    ('solid-shampoo',  'consumable', 'Solid Shampoo/Conditioner', 'Unbranded', 37, 0, '', 'Including waterproof bag', '{}', 'soap');

-- First aid
INSERT INTO _seed_items VALUES
    ('first-aid-kit-bag',  'first_aid', 'First aid kit bag',   'Unbranded', 22, 0, '', '', '{"kit_type":"kit"}', 'first aid kit'),
    ('tweezers-firstaid',  'first_aid', 'Tweezers',            'Unbranded',  5, 0, '', '', '{"kit_type":"tool"}', 'first aid kit'),
    ('sewing-kit',         'first_aid', 'Sewing Thread + Needle','Unbranded',2, 0, '', '', '{"kit_type":"tool"}', 'first aid kit'),
    ('blister-plaster',    'first_aid', 'Blister Pflaster',    'Compeed',   20, 0, '', '9 finger, 2 mid, 2 big size + case', '{"kit_type":"bandage","quantity":13}', 'first aid kit'),
    ('plasters',           'first_aid', 'Plasters',            'Unbranded', 15, 0, '', '6 regular, 2 square + case', '{"kit_type":"bandage","quantity":8}', 'first aid kit'),
    ('scissors',           'first_aid', 'Scissors',            'Unbranded',  7, 0, '', '', '{"kit_type":"tool"}', 'multi-tool'),
    ('elastic-plaster',    'first_aid', 'Elastic plaster strap','Unbranded', 9, 0, '', '', '{"kit_type":"bandage"}', 'first aid kit'),
    ('safety-pins',        'first_aid', '2 Safety pins',       'Unbranded',  1, 0, '', '', '{"kit_type":"tool","quantity":2}', 'first aid kit'),
    ('medical-tape',       'first_aid', 'Medical Tape',        'BSN',       20, 0, '', 'Leukotape P', '{"kit_type":"tape"}', 'first aid kit'),
    ('truetape',           'first_aid', 'Truetape',            'Truetape',  52, 0, '', '', '{"kit_type":"tape"}', 'first aid kit');

-- Wearables / clothing
INSERT INTO _seed_items VALUES
    ('socks-hiking-winter',   'footwear', 'Hiking Socks - Winter',  'Falke',     60, 21, 'https://www.falke.com', 'TK2 wool trekking socks', '{}', 'socks'),
    ('sleeping-shirt',        'wearable', 'Sleeping shirt',         'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'base layer'),
    ('sleeping-pants',        'wearable', 'Sleeping pants',         'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'base layer'),
    ('casual-tshirt',         'wearable', 'Casual - T-Shirt',       'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking shirt'),
    ('rain-jacket-eider',     'wearable', 'Rain jacket',            'Eider',    460,  0, 'https://www.eider.com/en/tonic-jkt-2-m-fire.html', 'Tonic jkt 2.0', '{"layer":"shell","waterproof":true}', 'rain jacket'),
    ('hiking-trousers-long',  'wearable', 'Hiking Trousers - Long', 'Eider',    320,  0, 'https://www.eider.com/en/flex-pant-m-agave-green.html', 'Flex Trousers', '{"layer":"base"}', 'hiking pants'),
    ('buff-winter',           'wearable', 'Buff - Winter',          'Buff',     116, 25, 'https://www.buff.com/de_de/merinowolle', 'Heavyweight Merino', '{"season":"winter","layer":"accessory"}', 'neck warmer'),
    ('gloves',                'wearable', 'Gloves',                 'Unbranded', 33,  0, '', '', '{"layer":"accessory"}', 'gloves'),
    ('cycling-socks-castelli','footwear', 'Cycling Socks - Summer', 'Castelli',  31,  0, '', '', '{}', 'socks'),
    ('underware',             'wearable', 'Underware',              'Unbranded', 56,  0, '', '', '{"layer":"base"}', 'base layer'),
    ('cycling-wind-jacket',   'wearable', 'Cycling Wind jacket',    'Vaude',    175,  0, '', 'Wind Jacket', '{"layer":"shell"}', 'rain jacket'),
    ('cycling-short-bibs',    'wearable', 'Cycling Short Bibs',     'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking shorts'),
    ('hiking-trousers-short', 'wearable', 'Hiking Trousers - Short','Unbranded',279,  0, '', '', '{"layer":"base"}', 'hiking shorts'),
    ('cycling-jersey',        'wearable', 'Cycling Jersey',         'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking shirt'),
    ('cycling-base-layer',    'wearable', 'Cycling Base Layer',     'BTWIN',     84,  0, '', '', '{"layer":"base"}', 'base layer'),
    ('buff-summer',           'wearable', 'Buff - Summer',          'Unbranded', 35,  0, '', '', '{"season":"summer","layer":"accessory"}', 'neck warmer'),
    ('cycling-rain-jacket',   'wearable', 'Cycling Rain Proof Jacket','Dynamics',210,55, 'https://shop.zweirad-stadler.de', 'Cycling Rain Jacket', '{"layer":"shell","waterproof":true}', 'rain jacket'),
    ('down-jacket',           'wearable', 'Down Jacket',            'Uniqlo',   236, 80, 'https://www.uniqlo.com', 'Ultralight Down Jacket', '{"layer":"mid"}', 'down jacket'),
    ('hiking-trousers-zip',   'wearable', 'Hiking Trousers - zipper','Schöffel',305,120,'https://www.schoeffel.com', 'Folkstone Zip Off', '{"layer":"base"}', 'hiking pants'),
    ('hiking-belt',           'wearable', 'Hiking Belt',            'Arcade',    49, 28, 'https://arcadebelts.com', 'Adventure Belt', '{"layer":"accessory"}', 'hiking pants'),
    ('poncho',                'wearable', 'Poncho',                 '3F UL Gear',197.5,45.5,'https://3fulgear.com/product/accessories/ultralight-tarp-poncho/', 'Ultralight Tarp Poncho', '{"layer":"shell","waterproof":true}', 'poncho'),
    ('thermal-leggings',      'wearable', 'Thermal underpants',     'Icebreaker',206,100,'https://www.icebreaker.com', 'Merino 200 Oasis Thermal Leggings', '{"season":"winter","layer":"base"}', 'base layer'),
    ('mid-layer-hoodie',      'wearable', 'Mid layer hoodie',       'Uniqlo',   250, 30, 'https://www.uniqlo.com', 'Ultra Stretch Dry-Ex Sweat Jacket', '{"layer":"mid"}', 'mid-layer'),
    ('casual-long-sleeve',    'wearable', 'Casual - long sleeve',   'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking shirt'),
    ('casual-short-trousers', 'wearable', 'Casual - Short trousers','Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking shorts'),
    ('casual-long-trousers',  'wearable', 'Casual - Long trousers', 'Unbranded', NULL, 0, '', '', '{"layer":"base"}', 'hiking pants'),
    ('swim-shorts',           'wearable', 'Swim Shorts',            'Unbranded',228, 26, '', '', '{"season":"summer","layer":"base"}', 'hiking shorts'),
    ('hiking-shirt-long',     'wearable', 'Hiking Shirt - Long',    'Uniqlo',   208,  0, '', '', '{"layer":"base"}', 'hiking shirt'),
    ('hiking-shirt-short',    'wearable', 'Hiking Shirt - Short',   'Unbranded',158,  0, '', '', '{"season":"summer","layer":"base"}', 'hiking shirt'),
    ('thermal-undershirt',    'wearable', 'Thermal undershirt',     'Icebreaker',220,100,'https://www.icebreaker.com', 'Merino 200 Oasis Long Sleeve', '{"season":"winter","layer":"base"}', 'base layer'),
    ('folding-cap',           'wearable', 'Folding Cap',            'Unbranded', 69, 13, 'https://www.amazon.de', '', '{"layer":"accessory"}', 'cap'),
    ('socks-summer-mckinley', 'footwear', 'Hiking Socks - Summer',  'McKinley',  47, 10, 'https://www.amazon.de', 'Quarter Socks', '{}', 'socks'),
    ('socks-summer-darntough','footwear', 'Hiking Socks - Summer',  'Darn Tough',60, 32, 'https://darntough.com', 'Micro Crew Lightweight Hiking Sock', '{}', 'socks'),
    ('cycling-socks-darntough','footwear','Cycling Socks - Summer', 'Darn Tough',42, 38, 'https://darntough.com', '1437 No Show Lightweight Athletic Sock', '{}', 'socks'),
    ('sleeping-socks',        'footwear', 'Sleeping socks',         'Unbranded', NULL, 0, '', '', '{}', 'socks');

-- Footwear
INSERT INTO _seed_items VALUES
    ('hiking-shoes-altra',     'footwear', 'Hiking Shoes',   'Altra Running', 608,144, 'https://www.altrarunning.eu/de/m-lone-peak-5.html', 'Lone Peak 5', '{"footwear_type":"trail_runner"}', 'trail runners'),
    ('cycling-shoes-scott',    'footwear', 'Cycling shoes',  'Scott',         892,  0, 'https://www.scott-sports.com', 'Crus-R', '{"footwear_type":"cycling"}', 'trail runners'),
    ('cycling-shoes-northwave','footwear', 'Cycling shoes',  'Northwave',     821,105, 'https://www.amazon.de', 'Origin Plus', '{"footwear_type":"cycling"}', 'trail runners'),
    ('casual-shoes',           'footwear', 'Casual shoes',   'VIVOBAREFOOT',  580, 90, 'https://www.vivobarefoot.com', 'Primus Knit EZ Men', '{"footwear_type":"casual"}', 'camp shoes'),
    ('hiking-sandals',         'footwear', 'Hiking Sandals', 'Xero',          319, 90, 'https://www.xeroshoes.eu/shop/sandals/ztrail-men/', 'Z-Trail EV', '{"footwear_type":"sandal"}', 'sandals');

-- Sleep system
INSERT INTO _seed_items VALUES
    ('sleeping-pillow',   'sleep', 'Sleeping pillow',   'Trekology',     110, 15, 'https://trekology.com', 'Luft 2.0', '{"fill_type":"air"}', 'pillow'),
    ('sleeping-bag-liner','sleep', 'Sleeping Bag liner','Sea to Summit', 248, 46, 'https://seatosummit.com/product/reactor-thermolite-mummy-liner/', 'Thermolite Reactor Liner', '{"fill_type":"synthetic"}', 'sleeping bag'),
    ('hammock',           'sleep', 'Hammock',           'Sea to Summit', 220, 85, 'https://seatosummit.com/product/hammock-set-ultralight-single/', 'Hammock Set Ultralight Single', '{}', 'bivy'),
    ('quilt',             'sleep', 'Quilt',             'Sea to Summit', 660,249, 'https://seatosummit.eu/products/cinder-down-quilt', 'Cinder Down Quilt Cd II Regular', '{"fill_type":"down","comfort_temp_c":2}', 'quilt'),
    ('sleeping-mat',      'sleep', 'Sleeping Mat',      'Sea to Summit', 595,170, 'https://seatosummit.eu/products/ultralight-insulated-mat', 'Ultralight Insulated Air — Large', '{"fill_type":"air","r_value":3.1}', 'sleeping pad'),
    ('down-blanket',      'sleep', 'Blanket',           'Get Out Gear',  460.7,77.2,'https://www.amazon.de', 'Down Puffy Blanket', '{"fill_type":"down"}', 'quilt'),
    ('silk-liner',        'sleep', 'Sleeping liner Silk','Sea to Summit',194,88.45,'https://www.amazon.de', 'Silk+cotton liner - Traveller', '{"fill_type":"other"}', 'sleeping bag');

-- Shelter
INSERT INTO _seed_items VALUES
    ('tent-lanshan2', 'shelter', 'Tent', '3F UL Gear', 1140, 200, 'https://www.amazon.de', 'LanShan 2 Person Tent — without poles', '{"capacity_people":2,"season_rating":"3-season","freestanding":false,"has_footprint":false}', 'tent');

-- Packs and bags
INSERT INTO _seed_items VALUES
    ('bike-rack',       'pack', 'Bike Rack',            'Racktime',      570,  0, 'https://www.racktime.com/en/racktime-products/system-carriers/racktime-product/lightit', 'light/t', '{"pack_type":"rack"}', 'pannier bag'),
    ('rack-small-bags', 'pack', 'Rack small bags',      'Arkel',         454,103, 'https://www.arkel-od.com/dry-lites/', 'Dry-lights', '{"pack_type":"pannier"}', 'pannier bag'),
    ('backpack-40-16',  'pack', 'Backpack',             '3F UL Gear',    900, 54, 'https://www.aliexpress.com/item/32980286902.html', '40+16L backpack', '{"pack_type":"backpack","capacity_liters":56}', 'backpack'),
    ('dry-bag-ortlieb', 'pack', 'Dry bag',              'Ortlieb',        54, 19, 'https://www.ortlieb.com/en_us/dry-bag-ps10+K20407', 'PS10 7L dry bag', '{"pack_type":"dry_bag","capacity_liters":7,"waterproof":true}', 'dry bag'),
    ('handlebar-bag',   'pack', 'Handlebar Bag',        'Topeak',        276, 55, 'https://www.topeak.com/global/en/products/189-Handlebar-Bags/1301-BARLOADER', 'BarLoader Handlebar Bag', '{"pack_type":"pannier"}', 'handlebar bag'),
    ('staff-sack',      'pack', 'Stuff sack',           'Sea to Summit',  40,  9, 'https://seatosummit.com', '6.5L Dry Bag', '{"pack_type":"stuff_sack","capacity_liters":6.5}', 'stuff sack'),
    ('dry-bag-35l',     'pack', 'Dry Bag 35L (Bag liner)','Sea to Summit',72, 30, 'https://seatosummit.eu/products/ultra-sil-dry-bag', 'Ultra Sil 35L', '{"pack_type":"dry_bag","capacity_liters":35,"waterproof":true}', 'dry bag'),
    ('pocket-bag',      'pack', 'Pocket bag (side bag)','Fjällräven',     80, 50, '', 'High Coast Pocket', '{"pack_type":"pocket"}', 'stuff sack'),
    ('backpack-talon33','pack', 'Backpack 33L',         'Osprey',       1011,150, 'https://www.ospreyeurope.com/shop/de_de/osprey-talon-33-2021', 'Talon 33', '{"pack_type":"backpack","capacity_liters":33}', 'backpack');

-- Camera
INSERT INTO _seed_items VALUES
    ('camera-tripod-pedco',  'camera', 'Camera tripod',           'Pedco',  51, 25, '', 'UltraPod 3', '{"camera_type":"tripod"}', 'camera'),
    ('camera-g5x',           'camera', 'Camera',                  'Canon', 350,  0, 'https://www.canon.com', 'G5x Mark II', '{"camera_type":"camera"}', 'camera'),
    ('camera-bag-tamrac',    'camera', 'Camera bag',              'Tamrac',105,  0, 'https://www.tamrac.com/products/pro-compact-1', 'Pro Compact 1', '{"camera_type":"case"}', 'camera'),
    ('action-camera-osmo',   'camera', 'Action camera',           'DJI',   155,235, 'https://www.dji.com/de/osmo-action', 'Osmo Action', '{"camera_type":"action_cam"}', 'camera'),
    ('osmo-action-case',     'camera', 'Osmo Action case',        'PGYTECH',229,13.29,'https://www.amazon.de', 'Mini Carrying Case', '{"camera_type":"case"}', 'camera'),
    ('camera-tripod-pgytech','camera', 'Camera tripod',           'PGYTECH', 95, 24, 'https://www.amazon.de', 'Mini Tripod Action Camera Tripod', '{"camera_type":"tripod"}', 'camera'),
    ('camera-hand-grip',     'camera', 'Camera Floating Hand Grip','Homeet',75.4,13, 'https://www.amazon.de', 'Floating Hand Grip', '{"camera_type":"accessory"}', 'camera'),
    ('tripod-phone-holder',  'camera', 'Tripod Phone Holder',     'Unbranded',54, 0, '', '', '{"camera_type":"accessory"}', 'camera');

-- Bike
INSERT INTO _seed_items VALUES
    ('mini-tool',       'bike', 'Mini tool',                 'Topeak',    164, 0, 'https://www.topeak.com', 'Ratchet Rocket Light DX', '{"bike_part":"tool"}', 'multi-tool'),
    ('lubricant',       'bike', 'Lubricant',                 'Unbranded', 150, 0, '', '', '{"bike_part":"accessory"}', 'multi-tool'),
    ('puncture-kit',    'bike', 'Puncture Repair Patch Kit', 'Unbranded',  30, 0, '', '', '{"bike_part":"accessory"}', 'multi-tool'),
    ('inner-tube',      'bike', 'Inner tube',                'BTWIN',     165, 0, '', '700X35-45 bike inner tube', '{"bike_part":"tube"}', 'multi-tool'),
    ('pressure-gauge',  'bike', 'Pressure Gauge',            'Schwalbe',   85, 0, 'https://www.schwalbe.com/Airmax-Pro-6010.01', 'Airmax Pro', '{"bike_part":"accessory"}', 'multi-tool'),
    ('hr-monitor',      'bike', 'HR monitor',                'Hammerhead',  22, 0, 'https://www.hammerhead.io', 'Heart Rate Monitor', '{"bike_part":"sensor"}', 'gps device'),
    ('cadence-sensor',  'bike', 'Cadence sensor',            'Garmin',     12,35, 'https://buy.garmin.com/de-DE/DE/p/641212', 'Cadence Sensor 2', '{"bike_part":"sensor"}', 'gps device'),
    ('speed-sensor',    'bike', 'Speed sensor',              'Garmin',     15,35, 'https://buy.garmin.com/de-DE/DE/p/641230', 'Speed Sensor 2', '{"bike_part":"sensor"}', 'gps device'),
    ('bike-lock',       'bike', 'Bike lock + mount',         'Abus',     1187, 0, 'https://www.abus.com', 'BORDO GRANIT XPlus 6500', '{"bike_part":"lock"}', 'multi-tool');

-- Electronics
INSERT INTO _seed_items VALUES
    ('kindle',             'electronics', 'Kindle',               'Amazon',          297,101, '', 'Kindle e-reader', '{"rechargeable":true,"charge_port":"micro-usb"}', 'notebook'),
    ('speaker-xb01',       'electronics', 'Speaker',              'Sony',            236,  0, 'https://www.sony.com/electronics/wireless-speakers/srs-xb01', 'XB01 Extra Bass', '{"rechargeable":true}', 'batteries'),
    ('charging-cable-x',   'electronics', 'Charging cable',       'Rolling Square',   80, 25, 'https://rollingsquare.com', 'InCharge X', '{"charge_port":"usb-c"}', 'batteries'),
    ('phone-iphone12',     'electronics', 'Phone',                'Apple',           164,  0, '', 'iPhone 12', '{"charge_port":"lightning","rechargeable":true}', 'gps device'),
    ('phone-case',         'electronics', 'Phone case',           'Spigen',           25,  0, '', 'iPhone 12 Case Thin Fit', '{}', 'gps device'),
    ('powerbank-5000',     'electronics', 'Powerbank small',      'Anker',           200, 14, 'https://www.anker.com/products/variant/powercore-5000/A1109031', 'PowerCore 5000', '{"capacity_mah":5000,"rechargeable":true,"charge_port":"micro-usb"}', 'power bank'),
    ('usbc-charger-nano',  'electronics', 'USB-C Charger',        'Anker',            35, 20, 'https://www.amazon.de', 'PowerPort III Nano', '{"charge_port":"usb-c"}', 'batteries'),
    ('bike-computer-karoo','electronics', 'Bike computer',        'Hammerhead',      132,  0, 'https://www.eu.hammerhead.io/pages/karoo2', 'Karoo 2', '{"rechargeable":true,"charge_port":"usb-c"}', 'gps device'),
    ('head-lamp-nu20',     'electronics', 'Head lamp',            'Nitecore',         47, 38, 'https://www.nitecore.com', 'NU20', '{"rechargeable":true,"charge_port":"usb-c"}', 'headlamp'),
    ('powerbank-20000-pd', 'electronics', 'Powerbank',            'Anker',           345, 46, 'https://www.amazon.de', 'PowerCore Essential 20000 PD', '{"capacity_mah":20000,"rechargeable":true,"charge_port":"usb-c"}', 'power bank'),
    ('headset-aeropex',    'electronics', 'Headset',              'AfterShokz',      103.3,118.95,'https://shokz.com', 'Aeropex', '{"rechargeable":true}', 'batteries'),
    ('airpods-pro',        'electronics', 'AirPods',              'Apple',            61,272, '', 'AirPods Pro', '{"rechargeable":true,"charge_port":"lightning"}', 'batteries'),
    ('charger-3port',      'electronics', '3 Port Charger USB-C', 'Anker',           131, 70, '', 'PowerPort III', '{"charge_port":"usb-c"}', 'batteries'),
    ('electronics-bag',    'electronics', 'Electronics Bag',      'Finpac',          168, 15, 'https://www.amazon.de', '', '{}', 'stuff sack'),
    ('powerbank-magsafe',  'electronics', 'Power bank 5k MagSafe','Anker',           134, 39, 'https://www.anker.com/eu-de/products/a1619', 'PowerCore Magnetic 5K', '{"capacity_mah":5000,"rechargeable":true,"charge_port":"usb-c"}', 'power bank'),
    ('apple-watch-8',      'electronics', 'Apple Watch 8',        'Apple',            50,589, '', 'Apple Watch 8', '{"rechargeable":true}', 'gps device'),
    ('apple-watch-cable',  'electronics', 'Apple Watch Charging Cable','Apple',       31, 29, '', '', '{}', 'batteries'),
    ('powerbank-prime-20k','electronics', 'Power bank 20k + bag', 'Anker',           555,  0, '', 'Anker Prime 20000 mAh', '{"capacity_mah":20000,"rechargeable":true,"charge_port":"usb-c"}', 'power bank'),
    ('usbc-charger-2port', 'electronics', 'USB C Charger 2 ports','Anker',            91,  0, 'https://www.amazon.de', '523 Nano III', '{"charge_port":"usb-c"}', 'batteries'),
    ('powerbank-nano-10k', 'electronics', 'Power bank 10K',       'Anker',           218,39.71,'https://www.amazon.de', 'Nano Power Bank', '{"capacity_mah":10000,"rechargeable":true,"charge_port":"usb-c"}', 'power bank'),
    ('lightning-cable',    'electronics', 'Lightning Cable',      'Apple',            19, 25, '', 'Lightning to USB-C Cable', '{"charge_port":"lightning"}', 'batteries'),
    ('keyboard',           'electronics', 'Keyboard',             'Unbranded',       NULL, 0, '', '', '{"rechargeable":true}', 'batteries'),
    ('mouse',              'electronics', 'Mouse',                'Unbranded',       NULL, 0, '', '', '{"rechargeable":true}', 'batteries'),
    ('ipad-pro',           'electronics', 'iPad',                 'Apple',           NULL, 0, '', 'iPad Pro', '{"rechargeable":true,"charge_port":"usb-c"}', 'notebook'),
    ('macbook-pro',        'electronics', 'MacBook',              'Apple',           NULL, 0, '', 'MacBook Pro', '{"rechargeable":true,"charge_port":"usb-c"}', 'notebook');

-- Toiletries (non-consumable hygiene)
INSERT INTO _seed_items VALUES
    ('clothesline',         'toiletries', 'Clothesline',     'Sea to Summit', 28,  9, 'https://seatosummit.com', 'The Clothesline', '{"toiletry_type":"other"}', 'clothesline'),
    ('towel-airlite',       'toiletries', 'Towel',           'Sea to Summit', 47, 23, 'https://seatosummit.com', 'Airlite L', '{"toiletry_type":"towel"}', 'towel'),
    ('towel-packtowl',      'toiletries', 'Towel',           'PackTowl',     290, 29, 'https://www.amazon.de', 'PackTowl Personal towel XXL', '{"toiletry_type":"towel"}', 'towel'),
    ('toothbrush',          'toiletries', 'Toothbrush',      'Unbranded',     18,  0, '', '', '{"toiletry_type":"brush"}', 'toothbrush'),
    ('nail-clippers',       'toiletries', 'Nail clippers',   'Unbranded',     15,  0, '', '', '{"toiletry_type":"tool"}', 'multi-tool'),
    ('soap-bag',            'toiletries', 'Soap bag',        'Matador',       11,  0, '', 'Bar Soap case', '{"toiletry_type":"bag"}', 'soap'),
    ('tweezers-toiletries', 'toiletries', 'Tweezers',        'Unbranded',      4,  0, '', '', '{"toiletry_type":"tool"}', 'multi-tool'),
    ('body-scrubber',       'toiletries', 'Body scrubber',   'Tooletries',    17,  0, 'https://tooletries.com.au/products/the-body-scrubber', '', '{"toiletry_type":"other"}', 'soap'),
    ('towel-cocoon',        'toiletries', 'Towel',           'Cocoon',       155, 20, 'https://www.cocoon.at/us/products/eco-travel-towel', 'Eco Travel Towel', '{"toiletry_type":"towel"}', 'towel'),
    ('toiletries-bag-deuter','toiletries','Toiletries bag',  'Deuter',        45, 22, 'https://www.deuter.com', 'Wash bag Tour II', '{"toiletry_type":"bag"}', 'stuff sack'),
    ('toiletries-bag-procase','toiletries','Toiletries bag', 'ProCase',       96,  0, 'https://www.amazon.de', '', '{"toiletry_type":"bag"}', 'stuff sack'),
    ('shaving-machine',     'toiletries', 'Shaving machine', 'Philips',      182.8,70, 'https://www.philips.de', 'OneBlade', '{"toiletry_type":"tool"}', 'multi-tool'),
    ('bite-away',           'toiletries', 'Bite away',       'Beurer',        49, 24, 'https://www.beurer.com', 'Insect bite healer', '{"toiletry_type":"tool"}', 'insect repellent');

-- Accessories / miscellaneous
INSERT INTO _seed_items VALUES
    ('ffp2-mask',          'accessory', 'FFP2 Mask',                'Unbranded',  3,  0, '', '', '{"subtype":"health"}', 'first aid kit'),
    ('sunglasses-case-duco','accessory','Sunglasses case',          'Duco',      57,  0, 'https://www.amazon.de', 'Polarised Sunglasses', '{"subtype":"eyewear"}', 'sunglasses'),
    ('picnic-blanket',     'accessory', 'Picnic blanket',           'Unbranded',750,  0, '', '', '{"subtype":"comfort"}', 'tarp'),
    ('sport-sunglasses',   'accessory', 'Sport Sunglasses',         'Alpina',    24, 38, 'https://www.amazon.de', 'Nylos Shield VL', '{"subtype":"eyewear"}', 'sunglasses'),
    ('umbrella',           'accessory', 'Umbrella',                 'Lejorain', 400,  0, 'https://www.amazon.de', '54" Umbrella', '{"subtype":"rain"}', 'poncho'),
    ('wallet',             'accessory', 'Wallet',                   'Unbranded',145,  0, '', '', '{"subtype":"documents"}', 'notebook'),
    ('folding-knife',      'accessory', 'Folding knife',            'Unbranded', 84,  0, '', '', '{"subtype":"tool"}', 'knife'),
    ('duct-tape',          'accessory', 'Duct Tape',                'Unbranded',NULL, 0, '', '', '{"subtype":"repair"}', 'duct tape'),
    ('hiking-sticks',      'accessory', 'Hiking sticks',            'Trekology',539, 35, 'https://www.amazon.de', 'Trek-Z', '{"subtype":"trekking"}', 'trekking poles'),
    ('bandana',            'accessory', 'Bandana',                  'Karl Loven',25,  2, 'https://www.amazon.de', '', '{"subtype":"apparel"}', 'neck warmer'),
    ('passport',           'accessory', 'Passport + BC',            'Unbranded', 36,  0, '', '', '{"subtype":"documents"}', 'notebook'),
    ('groceries-bag',      'accessory', 'Groceries Bag',            'Unbranded', 35,  0, '', '', '{"subtype":"storage"}', 'trash bag'),
    ('sunglasses-mykita',  'accessory', 'Sunglasses',               'Mykita',    16,  0, '', '', '{"subtype":"eyewear"}', 'sunglasses'),
    ('sunglasses-foldies', 'accessory', 'Sunglasses - folding',     'Foldies',   27, 75, 'https://www.foldies.com', 'Remix', '{"subtype":"eyewear"}', 'sunglasses'),
    ('sunglasses-foldies-case','accessory','Sunglasses - folding case','Foldies',40, 0, '', 'Vegan case', '{"subtype":"eyewear"}', 'sunglasses'),
    ('camino-guidebook',   'accessory', 'Camino del Norte guidebook','Wise Pilgrim',193,29,'https://shop.wisepilgrim.com', 'Wise Pilgrim Guide', '{"subtype":"navigation"}', 'map'),
    ('ear-plugs',          'accessory', 'Ear plugs',                'Loop',       7, 20, 'https://www.loopearplugs.com/products/quiet', 'Quiet', '{"subtype":"comfort"}', 'first aid kit'),
    ('phone-tablet-stand', 'accessory', 'Phone / Tablet stand',     'Unbranded',NULL, 0, '', '', '{"subtype":"stand"}', 'camera');

-- Resolve manufacturers and insert items + labels from the staging table.
INSERT INTO items (id, manufacturer_id, type_id, name, is_active, weight_grams, price, source_url, description, attributes_json)
SELECT md5('item-' || s.slug)::uuid,
       m.id,
       s.type_id,
       s.name,
       TRUE,
       s.weight_grams,
       NULLIF(s.price, 0),
       NULLIF(s.source_url, ''),
       NULLIF(s.description, ''),
       COALESCE(s.attributes, '{}'::jsonb)
FROM _seed_items s
JOIN manufacturers m ON lower(m.name) = lower(s.manufacturer)
ON CONFLICT (id) DO NOTHING;

INSERT INTO item_labels (item_id, label_id)
SELECT md5('item-' || s.slug)::uuid, l.id
FROM _seed_items s
JOIN labels l ON l.name = s.label
ON CONFLICT (item_id, label_id) DO NOTHING;

DROP TABLE _seed_items;

-- ---------------------------------------------------------------------------
-- Item sets: reusable groupings of items (kits) for building packs quickly.
-- Stable UUIDs via md5('set-'||slug) keep this idempotent. set_category
-- references an item_type that best describes the kit's theme.
-- ---------------------------------------------------------------------------
INSERT INTO item_sets (id, name, description, set_category)
VALUES
    (md5('set-sleep-system')::uuid,    'Sleep System',          'Pad, quilt, liner and pillow for a comfortable night.',          'sleep'),
    (md5('set-cook-kit')::uuid,        'Cook Kit',              'Stove, pot, fuel and eating utensils for camp meals.',           'cookware'),
    (md5('set-coffee-kit')::uuid,      'Trail Coffee Kit',      'Everything for fresh coffee on the trail.',                      'cookware'),
    (md5('set-water-system')::uuid,    'Water System',          'Filter and bottles for drinking water on the go.',               'water'),
    (md5('set-first-aid-kit')::uuid,   'First Aid Kit',         'Bandages, tape and tools for trail injuries.',                   'first_aid'),
    (md5('set-toiletries-kit')::uuid,  'Toiletries Kit',        'Hygiene essentials and quick-dry towel.',                        'toiletries'),
    (md5('set-rain-layer')::uuid,      'Rain Layer',            'Waterproof shell options for wet weather.',                      'wearable'),
    (md5('set-cold-layers')::uuid,     'Cold Weather Layers',   'Insulation and base layers for cold conditions.',                'wearable'),
    (md5('set-charging-kit')::uuid,    'Charging Kit',          'Power banks, chargers and cables to stay powered.',              'electronics'),
    (md5('set-camera-kit')::uuid,      'Photography Kit',       'Cameras, tripods and cases for capturing the trip.',             'camera'),
    (md5('set-bikepacking-kit')::uuid, 'Bikepacking Repair Kit','Tools, tube and lock for self-supported cycling.',               'bike'),
    (md5('set-bike-bags')::uuid,       'Bikepacking Bags',      'Racks and bags to carry gear on the bike.',                      'pack')
ON CONFLICT (id) DO NOTHING;

-- Staging table maps set membership by slug so it stays readable and in sync.
CREATE TEMP TABLE _seed_set_items (
    set_slug   TEXT,
    item_slug  TEXT,
    quantity   INT,
    sort_order INT
);

INSERT INTO _seed_set_items (set_slug, item_slug, quantity, sort_order) VALUES
    -- Sleep System
    ('set-sleep-system', 'sleeping-mat',       1, 10),
    ('set-sleep-system', 'quilt',              1, 20),
    ('set-sleep-system', 'sleeping-bag-liner', 1, 30),
    ('set-sleep-system', 'sleeping-pillow',    1, 40),

    -- Cook Kit
    ('set-cook-kit', 'stove',      1, 10),
    ('set-cook-kit', 'gas',        1, 20),
    ('set-cook-kit', 'pot-x',      1, 30),
    ('set-cook-kit', 'bowl-x',     1, 40),
    ('set-cook-kit', 'mug-x',      1, 50),
    ('set-cook-kit', 'spork',      1, 60),
    ('set-cook-kit', 'lighter',    1, 70),

    -- Trail Coffee Kit
    ('set-coffee-kit', 'coffee-maker',   1, 10),
    ('set-coffee-kit', 'coffee-grinder', 1, 20),
    ('set-coffee-kit', 'mug-x',          1, 30),

    -- Water System
    ('set-water-system', 'water-filter',         1, 10),
    ('set-water-system', 'water-bottle-vaude',   2, 20),
    ('set-water-system', 'soft-water-flask-500', 1, 30),

    -- First Aid Kit
    ('set-first-aid-kit', 'first-aid-kit-bag', 1, 10),
    ('set-first-aid-kit', 'blister-plaster',   1, 20),
    ('set-first-aid-kit', 'plasters',          1, 30),
    ('set-first-aid-kit', 'medical-tape',      1, 40),
    ('set-first-aid-kit', 'elastic-plaster',   1, 50),
    ('set-first-aid-kit', 'scissors',          1, 60),
    ('set-first-aid-kit', 'tweezers-firstaid', 1, 70),
    ('set-first-aid-kit', 'safety-pins',       1, 80),
    ('set-first-aid-kit', 'sewing-kit',        1, 90),

    -- Toiletries Kit
    ('set-toiletries-kit', 'toiletries-bag-deuter', 1, 10),
    ('set-toiletries-kit', 'toothbrush',            1, 20),
    ('set-toiletries-kit', 'toothpaste',            1, 30),
    ('set-toiletries-kit', 'solid-body-soap',       1, 40),
    ('set-toiletries-kit', 'towel-airlite',         1, 50),
    ('set-toiletries-kit', 'nail-clippers',         1, 60),
    ('set-toiletries-kit', 'sunscreen',             1, 70),

    -- Rain Layer
    ('set-rain-layer', 'rain-jacket-eider', 1, 10),
    ('set-rain-layer', 'poncho',            1, 20),
    ('set-rain-layer', 'umbrella',          1, 30),

    -- Cold Weather Layers
    ('set-cold-layers', 'down-jacket',        1, 10),
    ('set-cold-layers', 'mid-layer-hoodie',   1, 20),
    ('set-cold-layers', 'thermal-undershirt', 1, 30),
    ('set-cold-layers', 'thermal-leggings',   1, 40),
    ('set-cold-layers', 'buff-winter',        1, 50),
    ('set-cold-layers', 'gloves',             1, 60),
    ('set-cold-layers', 'socks-hiking-winter',2, 70),

    -- Charging Kit
    ('set-charging-kit', 'powerbank-20000-pd', 1, 10),
    ('set-charging-kit', 'charger-3port',      1, 20),
    ('set-charging-kit', 'usbc-charger-nano',  1, 30),
    ('set-charging-kit', 'charging-cable-x',   1, 40),
    ('set-charging-kit', 'lightning-cable',    1, 50),
    ('set-charging-kit', 'electronics-bag',    1, 60),

    -- Photography Kit
    ('set-camera-kit', 'camera-g5x',            1, 10),
    ('set-camera-kit', 'action-camera-osmo',    1, 20),
    ('set-camera-kit', 'camera-bag-tamrac',     1, 30),
    ('set-camera-kit', 'camera-tripod-pgytech', 1, 40),
    ('set-camera-kit', 'osmo-action-case',      1, 50),

    -- Bikepacking Repair Kit
    ('set-bikepacking-kit', 'mini-tool',      1, 10),
    ('set-bikepacking-kit', 'inner-tube',     1, 20),
    ('set-bikepacking-kit', 'puncture-kit',   1, 30),
    ('set-bikepacking-kit', 'pressure-gauge', 1, 40),
    ('set-bikepacking-kit', 'lubricant',      1, 50),
    ('set-bikepacking-kit', 'bike-lock',      1, 60),

    -- Bikepacking Bags
    ('set-bike-bags', 'bike-rack',       1, 10),
    ('set-bike-bags', 'rack-small-bags', 1, 20),
    ('set-bike-bags', 'handlebar-bag',   1, 30);

INSERT INTO set_items (set_id, item_id, quantity, sort_order)
SELECT md5(si.set_slug)::uuid,
       md5('item-' || si.item_slug)::uuid,
       si.quantity,
       si.sort_order
FROM _seed_set_items si
ON CONFLICT (set_id, item_id) DO NOTHING;

DROP TABLE _seed_set_items;

-- ---------------------------------------------------------------------------
-- Packing lists: named checklist templates tagged with existing labels.
-- Packing lists hold no items/sets directly (builder helpers); their labels
-- act as filters that pull the relevant gear into the list builder.
-- Stable UUIDs via md5('list-'||slug) keep this idempotent.
-- ---------------------------------------------------------------------------
INSERT INTO packing_lists (id, name, description)
VALUES
    (md5('list-ul-overnight')::uuid,    'Ultralight Overnight',     'Minimal kit for a single night in good weather.'),
    (md5('list-thru-hike')::uuid,       'Thru-Hike Essentials',     'Core gear for long-distance hiking with resupply.'),
    (md5('list-bikepacking')::uuid,     'Bikepacking Weekend',      'Self-supported cycling trip with repair and bags.'),
    (md5('list-winter-camp')::uuid,     'Winter Camping',           'Cold-weather layers and a warm sleep system.'),
    (md5('list-day-hike')::uuid,        'Day Hike',                 'Light essentials for a single-day outing.'),
    (md5('list-camino')::uuid,          'Camino Pilgrimage',        'Walking pilgrimage with town stays and light load.')
ON CONFLICT (id) DO NOTHING;

-- Staging table maps packing-list labels by slug + existing label name.
CREATE TEMP TABLE _seed_list_labels (
    list_slug  TEXT,
    label      TEXT
);

INSERT INTO _seed_list_labels (list_slug, label) VALUES
    -- Ultralight Overnight
    ('list-ul-overnight', 'tent'),
    ('list-ul-overnight', 'quilt'),
    ('list-ul-overnight', 'sleeping pad'),
    ('list-ul-overnight', 'stove'),
    ('list-ul-overnight', 'water filter'),
    ('list-ul-overnight', 'headlamp'),

    -- Thru-Hike Essentials
    ('list-thru-hike', 'backpack'),
    ('list-thru-hike', 'trail runners'),
    ('list-thru-hike', 'trekking poles'),
    ('list-thru-hike', 'water filter'),
    ('list-thru-hike', 'first aid kit'),
    ('list-thru-hike', 'power bank'),
    ('list-thru-hike', 'sunscreen'),

    -- Bikepacking Weekend
    ('list-bikepacking', 'pannier bag'),
    ('list-bikepacking', 'handlebar bag'),
    ('list-bikepacking', 'multi-tool'),
    ('list-bikepacking', 'gps device'),
    ('list-bikepacking', 'rain jacket'),

    -- Winter Camping
    ('list-winter-camp', 'tent'),
    ('list-winter-camp', 'sleeping bag'),
    ('list-winter-camp', 'sleeping pad'),
    ('list-winter-camp', 'base layer'),
    ('list-winter-camp', 'down jacket'),
    ('list-winter-camp', 'gloves'),
    ('list-winter-camp', 'neck warmer'),

    -- Day Hike
    ('list-day-hike', 'water bottle'),
    ('list-day-hike', 'snacks'),
    ('list-day-hike', 'first aid kit'),
    ('list-day-hike', 'sunglasses'),
    ('list-day-hike', 'rain jacket'),

    -- Camino Pilgrimage
    ('list-camino', 'trail runners'),
    ('list-camino', 'sandals'),
    ('list-camino', 'towel'),
    ('list-camino', 'sleeping bag'),
    ('list-camino', 'soap'),
    ('list-camino', 'map');

INSERT INTO packing_list_labels (packing_list_id, label_id)
SELECT md5(ll.list_slug)::uuid, l.id
FROM _seed_list_labels ll
JOIN labels l ON l.name = ll.label
ON CONFLICT (packing_list_id, label_id) DO NOTHING;

DROP TABLE _seed_list_labels;
