# Proposal: Refactor Arsitektur Form Node & Kanvas (Pemisahan Blueprint vs State)

## Alasan (Reason)
Saat ini, terjadi inkonsistensi struktur data *workflow* antara Backend, AI, dan UI. Backend dan AI menggunakan struktur JSON yang sangat ringan (hanya berisi *key-value* input). Namun, UI (Vue Flow) saat ini dirancang untuk menyimpan seluruh spesifikasi UI (seperti `auto_layout`, `label`, `tipe field`) secara "gabungan" ke dalam state `node.data` di memori kanvas. 

Hal ini memicu masalah fatal:
1. **Import/AI Form Blank**: Ketika UI mencoba membuka node yang diimpor dari file JSON atau hasil *generate* AI yang bentuknya minimalis (tanpa `auto_layout`), form gagal di-*render* (tampil kosong).
2. **Database Bloat**: Jika UI mencoba memaksakan perbaikan dengan me-*merge* `auto_layout` ke dalam state kanvas saat form dibuka, maka begitu tombol *Save* ditekan, seluruh metadata UI yang berukuran raksasa itu otomatis ikut terkirim ke *backend* dan mengotori *database*.

## Sebelum (Current State)
- Data node di dalam memori kanvas (`flowNodes.value[x].data.action`) membawa beban berlebih karena mengkloning keseluruhan spesifikasi katalog.
- *State* kanvas dipaksa menyimpan atribut kosmetik/visual seperti `auto_layout` dan atribut `component`.
- Komponen Form (seperti `EkAutoNodeForm.vue` dan `EkDynamicForm.vue`) sangat bergantung pada atribut layout yang terbenam di dalam instance `node.data` untuk bisa menggambar antarmuka.
- Alur *Save/Export* menjadi rentan karena membutuhkan pembersihan manual (*hacky code*) seperti perintah `delete result.auto_layout` untuk mencegah pencemaran payload.

## Sesudah (Proposed State)
- **State Kanvas Minimalis (Separation of Concerns)**: `node.data.action.fields` dikuruskan hingga murni hanya berfungsi menyimpan *Key* dan *Value* isian *user* (contoh: `{ "key": "url", "value": "google.com" }`).
- **Blueprint Real-time via Store**: Komponen Form akan menerima referensi layout (`auto_layout`) dan atribut komponen secara langsung dan *real-time* dari *Node Store (Catalog)*, bukan dari `node.data` kanvas.
- **Two-way Binding yang Tepat**: `<EkDynamicForm>` akan dipisahkan asupannya: Ia membaca *Blueprint* dari Store untuk urusan struktur/kosmetik (label, grid, tipe input), lalu mengikat nilai aslinya (`v-model`) murni pada array *Value* minimalis yang ada di Kanvas.

## Pros (Kelebihan)
- **100% Kebal Bug Import/Eksport**: Karena format data di dalam kanvas sama persis bentuknya dengan format JSON murni dari Backend/AI, tidak akan pernah ada lagi bentrok *missing layout* atau form kosong.
- **Database Super Bersih**: Tidak ada risiko *metadata* UI ikut tersimpan ke backend. *Payload* data menjadi sangat efisien dan ringan.
- **Pembaruan Layout Aman secara Global**: Jika *developer* kelak mengubah tata letak form di Katalog, seluruh *workflow* terdahulu akan langsung beradaptasi secara visual tanpa merusak *value* asli atau butuh migrasi basis data.

## Cons (Kekurangan)
- **Waktu Refactoring Awal**: Memerlukan sedikit *refactoring* fundamental, terutama mengubah struktur `props` pada `EkDynamicForm.vue` agar mendukung pemisahan struktur dan *value*.
- **Pembersihan Data Lama**: Apabila terdapat *workflow* lama di database yang terlanjur kotor membawa `auto_layout` akibat arsitektur sebelumnya, secara teknis UI baru akan mengabaikannya, namun data itu tetap memberatkan database jika tidak dibersihkan (*scrubbing*).

## Review Teknis

### Verdict
Proposal ini **penting dan layak diprioritaskan**. Masalahnya bukan sekadar kosmetik UI, tetapi kontrak data yang saat ini mencampur tiga hal berbeda:
1. **Catalog/blueprint**: definisi schema, label, component, layout, default, dan output contract.
2. **Workflow instance**: pilihan action dan nilai yang diisi user.
3. **Canvas state**: posisi node, state editor, dan metadata tampilan sementara.

Selama tiga lapisan ini tetap bercampur, bug import/export, payload membesar, dan risiko data lama pecah saat catalog berubah akan terus muncul.

### Bukti dari Codebase Saat Ini
- `EkAutoNodeForm.vue` merender form dari `localAction.auto_layout`. Jika action hasil import tidak membawa `auto_layout`, layout menjadi kosong walaupun catalog sebenarnya punya blueprint.
- `buildSavePayload()` menyimpan `n.data.action` apa adanya ke workflow. Jika `action` di kanvas berisi `auto_layout`, metadata UI ikut terkirim dan tersimpan.
- `buildActionInstance()` meng-clone action definition dari catalog, termasuk metadata action dan field schema, lalu menambahkan value. Ini akar dari pencampuran blueprint dan instance.
- Backend executor sebagian besar hanya membutuhkan `action.key`, `action.response_var`, dan `fields[].value`, tetapi validator Go saat ini masih membandingkan `fields[].type` dan `fields[].required` dari payload user terhadap catalog.

### Koreksi Penting untuk Proposal
Bagian "State Kanvas Minimalis" benar, tetapi belum cukup jika hanya diterapkan di UI. Backend sekarang belum sepenuhnya siap menerima `fields` yang hanya berisi `{ key, value }`.

Jika payload field dibuat minimalis sekarang:

```json
{ "key": "url", "value": "https://example.com" }
```

maka validasi Go berisiko gagal karena `type` kosong dan `required` default `false`, sementara catalog mungkin punya `type: "string"` dan `required: true`.

Jadi perubahan harus mencakup salah satu dari dua pendekatan ini:

1. **Canonical minimal payload + hydrate saat validasi/run**.
   Backend menerima field instance minimal, lalu mengambil schema dari catalog berdasarkan `node.type`, `action.key`, dan `field.key`.

2. **Transitional payload sanitizer di frontend**.
   UI tetap menyimpan state minimal, tetapi sebelum save mengirim field yang sudah di-hydrate dari catalog. Ini lebih cepat, tetapi kurang bersih karena backend masih bergantung pada payload schema dari client.

Rekomendasi saya: pilih opsi pertama sebagai target akhir. Backend adalah tempat paling tepat untuk memastikan schema canonical, karena AI/import/API tidak selalu lewat UI.

### Bentuk Data yang Disarankan
Target workflow instance sebaiknya seperti ini:

```json
{
  "id": "abc12",
  "type": "http",
  "label": "HTTP Request",
  "action": {
    "key": "request",
    "response_var": "http.request_x7k2m",
    "fields": [
      { "key": "url", "value": "https://example.com" },
      { "key": "method", "value": "GET" }
    ]
  }
}
```

Yang tidak boleh tersimpan di workflow instance:
- `auto_layout`
- `component`
- `label` field
- `description` action
- `options` field UI
- `required` field
- `type` field
- default value dari catalog kecuali memang sudah menjadi value eksplisit user

### Tambahan yang Kurang di Proposal

#### 1. Hydration Layer
Tambahkan fungsi eksplisit untuk menggabungkan blueprint dan value tanpa mengubah persisted state.

Contoh tanggung jawab:
- `getActionBlueprint(nodeType, actionKey)` mengambil action definition dari catalog.
- `hydrateActionForForm(instanceAction, blueprintAction, globalFields)` menghasilkan object render-only untuk form.
- `serializeActionForSave(action)` membuang metadata UI/schema sebelum save.

#### 2. Contract Test
Perlu test yang mengunci perilaku utama:
- Import workflow minimal dari AI tetap menampilkan form.
- Save payload tidak mengandung `auto_layout`.
- Backend validate menerima field minimal `{ key, value }`.
- Unknown field key ditolak atau diabaikan secara eksplisit, jangan ambigu.
- Existing dirty workflow yang punya `auto_layout` tetap bisa dibuka dan ketika disimpan ulang metadata tersebut hilang.

#### 3. Migration/Scrubbing
Jangan hanya "UI mengabaikan". Tambahkan mekanisme pembersihan:
- lazy cleanup saat workflow disimpan ulang, dan/atau
- command/helper satu kali untuk membersihkan workflow lama.

Target scrubbing:
- hapus `action.auto_layout`;
- hapus metadata schema dari `action.fields`;
- pertahankan `key`, `value`, dan field instance yang memang runtime-specific.

#### 4. MyNodes dan Custom Form
Proposal perlu menyebut dampak ke `mynodes` dan custom form seperti `HttpNodeForm.vue`. Jangan hanya fokus `EkAutoNodeForm.vue`, karena node khusus bisa punya serializer sendiri dan tetap menyimpan field schema.

### Alur Data End-to-End yang Harus Dijaga

#### 1. Load Catalog
`WorkflowEditor.vue` sudah memanggil `nodeStore.loadCatalog()` sebelum `editor.loadWorkflow()`. Ini penting dan harus tetap dijaga, karena workflow minimal hanya bisa dirender jika catalog sudah tersedia untuk hydration.

#### 2. Load Workflow / Import JSON
Workflow dari backend/import sebaiknya tetap minimal. Saat masuk ke Vue Flow, `mapNodesToFlow()` + `buildNodeData()` bertugas membuat `node.data` yang kaya metadata untuk kebutuhan UI:
- `label`, `icon`, `tags` dari workflow jika custom, fallback ke catalog;
- `nodeType` dari `node.type`;
- `outputs` dari catalog;
- action instance tetap minimal, tetapi informasi action blueprint harus bisa dicari dari catalog berdasarkan `nodeType + action.key`.

Dengan pola ini, `NodeCard.vue` tetap mendapat data yang dibutuhkan tanpa workflow menyimpan `auto_layout` atau schema field.

#### 3. Render NodeCard
`NodeCard.vue` saat ini butuh:
- `data.label`;
- `data.icon`;
- `data.tags`;
- `data.nodeType`;
- `data.action.key`;
- `data.outputs`;
- `data.sourceType`;
- `data.name`.

Semua ini bisa dipenuhi dari hydrated `node.data`. Yang perlu dihindari adalah membuat `NodeCard` bergantung pada `action.auto_layout`, `action.fields[].label`, atau metadata field lain dari workflow instance.

#### 4. Buka Form
`BaseNodeForm.vue` saat ini menentukan `currentActionHasOutput` dari `action.has_response`. Jika action instance dibuat minimal, nilai ini harus pindah ke catalog action blueprint. Jika tidak, field `Result name` bisa hilang untuk action yang sebenarnya punya response.

`EkAutoNodeForm.vue` juga harus mencari action blueprint dari catalog, lalu memisahkan:
- blueprint action/fields/layout untuk render;
- action values dari workflow instance untuk binding.

#### 5. Custom Form HTTP
`HttpNodeForm.vue` saat ini mengambil value dari `props.node.data.action.fields`, tetapi saat `getData()` ia meng-clone `props.node.data.action` dan hanya meng-update fields yang sudah ada. Jika action minimal atau dirty, behavior-nya bisa tidak konsisten.

Custom form harus punya helper yang sama:
- baca field blueprint dari catalog;
- baca value dari action instance minimal;
- return action minimal saat save.

#### 6. Save / Export
`buildSavePayload()` sekarang menyimpan `n.data.action` apa adanya. Ini harus menjadi titik serializer utama:
- hapus `auto_layout`;
- hapus field schema metadata;
- simpan hanya `action.key`, `action.response_var`, dan `fields[{ key, value }]`;
- jangan ikut menyimpan `outputs`, `action_has_response`, atau metadata render lain.

#### 7. MyNodes
`mynodeStore.saveItem()` juga menyimpan `nodeData.action` apa adanya. Kalau workflow save sudah bersih tetapi MyNodes tidak, metadata UI tetap bisa bocor lewat template yang disimpan ulang. MyNodes perlu memakai serializer yang sama dengan workflow save.

#### 8. UI Validation
`validateNodeConfig()` sekarang memvalidasi `action.fields` sebagai self-describing schema. Setelah action minimal, validator UI harus memvalidasi value instance terhadap catalog fields, bukan terhadap metadata di payload.

#### 9. Backend Validation / Runtime
Backend validator perlu menjadi sumber kebenaran:
- action key divalidasi terhadap catalog;
- required/type/options divalidasi dari catalog;
- field instance minimal `{ key, value }` diterima;
- unknown field key harus diputuskan: ditolak untuk strictness, atau diabaikan dengan alasan kompatibilitas.

Runtime executor sudah relatif cocok dengan payload minimal karena mayoritas node membaca value lewat `node.FieldValue(action, key)`. Namun default value sebaiknya di-hydrate dari catalog sebelum eksekusi, bukan bergantung pada `field.default` yang dikirim client.

### Prioritas Implementasi yang Disarankan
1. Ubah `EkAutoNodeForm.vue` agar layout dan field blueprint dibaca dari catalog, bukan dari `localAction`.
2. Ubah `EkDynamicForm.vue` agar menerima `layout` dan `fields` sebagai blueprint, serta value map/field values sebagai model terpisah.
3. Tambahkan serializer frontend yang memastikan payload save minimal dan membuang `auto_layout`.
4. Ubah backend validation agar field instance minimal divalidasi terhadap catalog tanpa mewajibkan client mengirim `type` dan `required`.
5. Tambahkan test import minimal, save minimal, dan dirty workflow cleanup.
6. Setelah stabil, jalankan scrubbing data lama.

### Risiko Jika Tidak Dikerjakan
- Workflow hasil AI/import akan terus rawan tidak bisa diedit dari UI.
- Database akan menyimpan metadata UI yang bisa membesar dan berubah-ubah.
- Perubahan catalog di masa depan bisa menimbulkan inkonsistensi antara workflow lama dan definisi node terbaru.
- Validasi/debug akan semakin sulit karena sulit membedakan mana schema resmi dan mana salinan lama di workflow instance.

### Kesimpulan
Proposal ini **penting**, tetapi perlu dinaikkan levelnya dari refactor UI menjadi **perbaikan kontrak data end-to-end**. Pemisahan blueprint dan state harus berlaku di UI, save/export, import, backend validation, dan runtime hydration. Tanpa bagian backend dan serializer, proposal ini hanya memindahkan sumber bug dari form kosong ke payload/validasi.
