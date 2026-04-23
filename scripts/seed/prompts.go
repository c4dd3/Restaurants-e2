package main

// prompts.go — templates de prompts para el LLM.
//
// Cada función devuelve (systemPrompt, userPrompt string). Se mantienen como
// funciones (no constantes) para que se puedan parametrizar (ej: cantidad N).
//
// ───────────────────────────────────────────────────────────────────────────
// 1. Restaurantes
// ───────────────────────────────────────────────────────────────────────────
//
//   promptRestaurants(n int) (system, user string)
//
//   system = "Eres un generador de datos sintéticos para una app de
//            restaurantes en Costa Rica. Devuelves SIEMPRE JSON válido,
//            sin markdown. Los nombres y direcciones deben ser realistas
//            pero ficticios."
//
//   user = fmt.Sprintf(`Genera %d restaurantes con la siguiente forma:
//     {
//       "restaurants": [
//         {
//           "name": string (2-4 palabras, creativo pero plausible),
//           "description": string (1-2 oraciones),
//           "address": string (calle + provincia de Costa Rica),
//           "category": one of ["mexicana","italiana","tica","japonesa","mariscos","vegetariana","parrilla","postres","cafetería"]
//         }
//       ]
//     }
//     IMPORTANTE: nombres NO repetidos. Variedad de provincias.`, n)
//
// ───────────────────────────────────────────────────────────────────────────
// 2. Menús y productos (combinado en un solo prompt para coherencia)
// ───────────────────────────────────────────────────────────────────────────
//
//   promptMenuWithProducts(restaurantName, restaurantCategory string, numProducts int)
//
//   user = fmt.Sprintf(`El restaurante "%s" es de categoría "%s".
//     Genera UN menú con %d productos coherentes con esa categoría.
//     Forma:
//     {
//       "menu": {
//         "name": string (ej: "Menú principal", "Especiales del chef"),
//         "products": [
//           { "name": str, "description": str, "price": number (en CRC, 2000-15000),
//             "category": str (entrada|plato fuerte|bebida|postre) }
//         ]
//       }
//     }`, restaurantName, restaurantCategory, numProducts)
//
// ───────────────────────────────────────────────────────────────────────────
// 3. Usuarios
// ───────────────────────────────────────────────────────────────────────────
//
//   promptUsers(n int)
//
//   user = `Genera N usuarios con:
//     {"users":[{"name": str (nombre + apellido tico/latino), "email": str (formato realista, no repetido)}]}
//     Contraseñas NO se generan — se asignarán todas como "password123" en el script.`
//
// ───────────────────────────────────────────────────────────────────────────
// 4. (Opcional) Reviews / comentarios — fuera de Etapa 2.
// ───────────────────────────────────────────────────────────────────────────
//
// Tips:
//   - Usar temperature ~0.8 para creatividad en nombres; 0.3 si se quiere
//     reproducibilidad.
//   - Incluir few-shot examples en el system prompt mejora notablemente
//     la coherencia de los precios y direcciones.
//   - Para evitar homogeneidad, incluir seed/nonce en cada prompt
//     (ej: "batch-id: 7") — el LLM varía más con eso.
