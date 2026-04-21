#!/usr/bin/env bash
# ============================================================================
# Inicialización del cluster MongoDB para Restaurants-e2.
#
# Este script se ejecuta una sola vez cuando se levanta el stack con el perfil mongo.
# Hace lo siguiente:
#   1. Inicia el replica set del config server.
#   2. Inicia los replica sets de shard1 y shard2 (cada uno con 3 nodos: 1 primario + 2 secundarios).
#   3. Agrega ambos shards al cluster vía mongos.
#   4. Habilita sharding en la base de datos "restaurants".
#   5. Define shard keys:
#        - products    → key "category" (hashed)    — distribuye por categoría
#        - reservations → key "restaurant_id" (hashed) — colocalidad por restaurante
#
# El script es idempotente: si ya fue inicializado, los comandos fallarán
# con "already initialized" y se continúa. Se puede correr múltiples veces sin daño.
# ============================================================================
set -e

echo "[mongo-init] esperando a que los nodos arranquen..."
sleep 15

MONGO="mongosh --quiet"

echo "[mongo-init] iniciando replica set del config server (cfgrs)..."
$MONGO --host mongo_configsvr:27019 --eval '
  try {
    rs.initiate({
      _id: "cfgrs",
      configsvr: true,
      members: [{ _id: 0, host: "mongo_configsvr:27019" }]
    });
  } catch (e) { print("cfgrs ya inicializado: " + e.message); }
'

echo "[mongo-init] iniciando replica set shard1rs (1 primario + 2 secundarios)..."
$MONGO --host mongo_shard1_a:27018 --eval '
  try {
    rs.initiate({
      _id: "shard1rs",
      members: [
        { _id: 0, host: "mongo_shard1_a:27018", priority: 2 },
        { _id: 1, host: "mongo_shard1_b:27018", priority: 1 },
        { _id: 2, host: "mongo_shard1_c:27018", priority: 1 }
      ]
    });
  } catch (e) { print("shard1rs ya inicializado: " + e.message); }
'

echo "[mongo-init] iniciando replica set shard2rs (1 primario + 2 secundarios)..."
$MONGO --host mongo_shard2_a:27018 --eval '
  try {
    rs.initiate({
      _id: "shard2rs",
      members: [
        { _id: 0, host: "mongo_shard2_a:27018", priority: 2 },
        { _id: 1, host: "mongo_shard2_b:27018", priority: 1 },
        { _id: 2, host: "mongo_shard2_c:27018", priority: 1 }
      ]
    });
  } catch (e) { print("shard2rs ya inicializado: " + e.message); }
'

echo "[mongo-init] esperando a que los replica sets elijan primario..."
sleep 10

echo "[mongo-init] agregando shards al cluster vía mongos..."
$MONGO --host mongos:27017 --eval '
  try { sh.addShard("shard1rs/mongo_shard1_a:27018,mongo_shard1_b:27018,mongo_shard1_c:27018"); }
  catch (e) { print("shard1 ya agregado: " + e.message); }
  try { sh.addShard("shard2rs/mongo_shard2_a:27018,mongo_shard2_b:27018,mongo_shard2_c:27018"); }
  catch (e) { print("shard2 ya agregado: " + e.message); }
'

echo "[mongo-init] habilitando sharding en la BD restaurants y definiendo shard keys..."
$MONGO --host mongos:27017 --eval '
  try { sh.enableSharding("restaurants"); } catch (e) { print(e.message); }

  // products: shard por categoría — distribuye escrituras por categoría culinaria.
  try {
    db.getSiblingDB("restaurants").products.createIndex({ category: "hashed" });
    sh.shardCollection("restaurants.products", { category: "hashed" });
  } catch (e) { print("products ya shardeado: " + e.message); }

  // reservations: shard por restaurant_id — consultas de disponibilidad filtran por restaurante,
  // así que los datos de un mismo restaurante quedan colocados.
  try {
    db.getSiblingDB("restaurants").reservations.createIndex({ restaurant_id: "hashed" });
    sh.shardCollection("restaurants.reservations", { restaurant_id: "hashed" });
  } catch (e) { print("reservations ya shardeado: " + e.message); }
'

echo "[mongo-init] estado final del cluster:"
$MONGO --host mongos:27017 --eval 'sh.status()'

echo "[mongo-init] OK — cluster listo."
