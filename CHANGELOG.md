# Changelog

## [1.0.2] - 2025-07-05

### Modifié
- **Endpoint requis** : Le paramètre `endpoint` est maintenant obligatoire au lieu d'être optionnel
- **Provider générique** : Le provider est maintenant compatible avec n'importe quelle API DNS respectant le format des routes


## [1.0.1] - 2025-06-30

### Ajouté
- **Support complet de libdns v1.0.0** : Migration vers les nouvelles structures de données
- **Nouvelles structures spécifiques** :
  - `libdns.Address` pour les enregistrements A/AAAA avec champ `IP` de type `netip.Addr`
  - `libdns.TXT` pour les enregistrements TXT avec champ `Text`
  - `libdns.CNAME` pour les enregistrements CNAME avec champ `Target`
  - `libdns.MX` pour les enregistrements MX avec champs `Preference` et `Target`
  - `libdns.NS` pour les enregistrements NS avec champ `Target`
- **Parsing intelligent des enregistrements MX** : Support des formats "10 mail.example.com" et "mail.example.com"
- **Validation des adresses IP** : Utilisation de `netip.ParseAddr` pour valider les adresses A/AAAA
- **Champ `ProviderData`** : Support du nouveau champ `ProviderData any` dans toutes les structures

### Modifié
- **Méthode `GetRecords`** : Retourne maintenant les structures spécifiques au lieu de `libdns.RR`
- **Méthode `convertToSpecificTypes`** : Améliorée pour utiliser les nouvelles structures
- **Fichier de test** : Mis à jour pour utiliser les nouvelles structures libdns v1.0.0
- **README** : Documentation mise à jour avec exemples des nouvelles structures

### Supprimé
- **Dépendance sur les anciennes structures** : Plus d'utilisation directe de `libdns.RR` pour les types supportés

### Compatibilité
- **Breaking Change** : Ce provider nécessite maintenant libdns v1.0.0 ou supérieur
- **Migration** : Les utilisateurs doivent migrer vers les nouvelles structures de données

## [1.0.0] - 2024-06-27

### Ajouté
- Support initial des interfaces libdns
- Implémentation des méthodes `GetRecords`, `AppendRecords`, `SetRecords`, `DeleteRecords`
- Support des validations ACME pour Caddy
- Gestion des enregistrements DNS via l'API immosquare 