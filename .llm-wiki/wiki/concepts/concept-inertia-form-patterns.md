---
type: concept
status: active
---

# Inertia Form Patterns (Create / Update Data)

Convention for CRUD form submissions in Laju Go using Inertia.js + Svelte 5.

Related: [[concept-inertia-spa-navigation]] (navigation vs form CRUD), [[concept-csrf-protection]]

## Decision Rule

| Need | Use |
|------|-----|
| Simple form — just collect data and submit | `<Form>` component |
| Pre-submit validation (check password match, etc.) | `useForm` + `<form>` |
| `fetch()` integration (avatar upload then save URL) | `useForm` + `<form>` |
| Reactive `bind:value` (live preview, dirty tracking) | `useForm` + `<form>` |
| Programmatic submit (trigger from outside form) | `useForm` + `<form>` |

**Default:** if unsure, use `useForm` + `<form>` — it handles more cases.

## Pattern A: `<Form>` Component — Simple Forms

No `bind:` needed — Form collects data from `name` attributes. Least boilerplate.

```svelte
<script lang="ts">
  import { Form } from '@inertiajs/svelte'
</script>

<!-- CREATE -->
<Form action="/users" method="post">
  <input type="text" name="name" />
  <input type="email" name="email" />
  <button type="submit">Create User</button>
</Form>

<!-- UPDATE -->
<Form action="/users/{user.id}" method="put">
  <input type="text" name="name" value={user.name} />
  <button type="submit">Update</button>
</Form>
```

Slot props for errors/processing (Svelte 5 snippet syntax):

```svelte
<Form action="/users" method="post">
  {#snippet children({ errors, processing, wasSuccessful })}
    <input type="text" name="name" />
    {#if errors.name}<div>{errors.name}</div>{/if}
    <button disabled={processing}>
      {processing ? 'Creating...' : 'Create User'}
    </button>
    {#if wasSuccessful}<div>Created!</div>{/if}
  {/snippet}
</Form>
```

## Pattern B: `useForm` + `<form>` — Forms with Validation/Control

Auto-tracks `processing`, `errors`, `isDirty`, `wasSuccessful`. Allows pre-submit validation and `fetch()` integration.

### Create

```svelte
<script lang="ts">
  import { useForm } from '@inertiajs/svelte'

  const form = useForm({
    name: '',
    email: '',
    role: 'user',
  })

  function submit(e: Event) {
    e.preventDefault()
    form.post('/users', {
      onSuccess: () => form.reset(),
    })
  }
</script>

<form onsubmit={submit}>
  <input bind:value={form.name} />
  {#if form.errors.name}<span>{form.errors.name}</span>{/if}
  <button disabled={form.processing}>Create</button>
</form>
```

### Update

```svelte
<script lang="ts">
  import { useForm } from '@inertiajs/svelte'

  let { user } = $props()

  const form = useForm(`EditUser:${user.id}`, {
    name: user.name,
    email: user.email,
  })

  function submit(e: Event) {
    e.preventDefault()
    form.put(`/users/${user.id}`)
  }
</script>

<form onsubmit={submit}>
  <input bind:value={form.name} />
  {#if form.isDirty}<span>Unsaved changes</span>{/if}
  <button disabled={form.processing}>Save</button>
</form>
```

### File Upload + Form Save (two-step)

Avatar upload uses `fetch() + FormData` for the file, then `form.put()` to persist the URL:

```svelte
<script lang="ts">
  import { useForm } from '@inertiajs/svelte'
  import { getCSRFToken } from '@lib/utils/csrf'

  const form = useForm('EditProfile', {
    name: user?.name ?? '',
    avatar: user?.avatar ?? '',
  })

  function handleAvatarChange(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('file', file)
    fetch('/app/upload', {
      method: 'POST',
      headers: { 'X-XSRF-TOKEN': getCSRFToken() },
      body: formData,
    })
      .then(r => r.json())
      .then(data => {
        if (data.success && data.url) {
          form.avatar = data.url
          form.put('/app/profile')
        }
      })
  }
</script>
```

## Key Rules (both patterns)

| Rule | Why |
|------|-----|
| `form.post()` for create, `form.put()`/`form.patch()` for update | Correct HTTP method, server knows intent |
| Unique key for edit forms: `useForm('EditUser:${id}', data)` | Persists form data + errors to history state |
| `disabled={form.processing}` or `disabled={processing}` | Prevent double-submit |
| `form.errors.field` or `errors.field` | Server validation errors auto-populate |
| `e.preventDefault()` in `useForm` submit handler | Prevents full page reload — Inertia sends XHR instead |
| File upload: `fetch() + FormData` then `form.put()` | Inertia forms can't send files directly |

## Sources

- [Inertia.js Forms Documentation (v3)](https://inertiajs.com/docs/v3/the-basics/forms)
- [Inertia.js Manual Visits (v3)](https://inertiajs.com/docs/v3/the-basics/manual-visits)
- [Inertia.js Svelte 5 Playground — Form.svelte](https://github.com/inertiajs/inertia/blob/2f33ed08/playgrounds/svelte5/resources/js/Pages/Form.svelte)
