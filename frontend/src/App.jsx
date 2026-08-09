import React, { useEffect, useMemo, useState } from 'react'
import { api, authStore, login, logout } from './api.js'

const icons = {
  dashboard: '▦', users: '◉', roles: '◆', permissions: '⌘', audit: '◎', sessions: '◷', logout: '↪', shield: '⬢'
}
const actorLabels = { USER: 'Usuario', INSTRUCTOR: 'Instructor', LEARNER: 'Aprendiz' }
const outcomeLabels = { SUCCESS: 'Exitoso', INVALID_PASSWORD: 'Clave incorrecta', USER_NOT_FOUND: 'Usuario no existe', ACCOUNT_LOCKED: 'Cuenta bloqueada', TOKEN_EXPIRED: 'Token expirado' }

function useHashPage() {
  const read = () => location.hash.replace('#/', '') || 'dashboard'
  const [page, setPage] = useState(read)
  useEffect(() => { const fn = () => setPage(read()); addEventListener('hashchange', fn); return () => removeEventListener('hashchange', fn) }, [])
  return [page, p => { location.hash = `#/${p}` }]
}

function App() {
  const [user, setUser] = useState(null), [checking, setChecking] = useState(true)
  useEffect(() => {
    if (!authStore.access()) { setChecking(false); return }
    api('/auth/me').then(setUser).catch(() => authStore.clear()).finally(() => setChecking(false))
  }, [])
  if (checking) return <Splash />
  if (!user) return <Login onLogin={setUser} />
  return <Shell user={user} setUser={setUser} />
}

function Splash() { return <div className="center-screen"><div className="spinner" /><p>Cargando IAM...</p></div> }

function Login({ onLogin }) {
  const [email, setEmail] = useState('admin@sena.edu.co'), [password, setPassword] = useState('Admin123*'), [error, setError] = useState(''), [busy, setBusy] = useState(false)
  async function submit(e) { e.preventDefault(); setBusy(true); setError(''); try { onLogin(await login(email, password)) } catch (e) { setError(e.message) } finally { setBusy(false) } }
  return <main className="login-page">
    <section className="login-brand">
      <div className="brand-mark">IAM</div><div><span className="eyebrow light">Sistema de identidad</span><h1>Acceso seguro.<br/>Permisos claros.</h1><p>Administración centralizada de identidades, roles, permisos y sesiones sobre el modelo RBAC de la base de datos.</p></div>
      <div className="brand-points"><span>✓ PostgreSQL 16</span><span>✓ RBAC por features</span><span>✓ Auditoría de accesos</span></div>
    </section>
    <section className="login-panel"><form className="login-card" onSubmit={submit}>
      <div className="mini-brand"><span className="logo-dot">I</span><strong>IAM Console</strong></div><h2>Bienvenido</h2><p className="muted">Ingresa con tus credenciales para continuar.</p>
      {error && <div className="alert error">{error}</div>}
      <label>Correo electrónico<input type="email" value={email} onChange={e=>setEmail(e.target.value)} required /></label>
      <label>Contraseña<input type="password" value={password} onChange={e=>setPassword(e.target.value)} required /></label>
      <button className="btn primary wide" disabled={busy}>{busy ? 'Ingresando...' : 'Iniciar sesión'}</button>
      <div className="demo-box"><b>Credenciales de demostración</b><code>admin@sena.edu.co</code><code>Admin123*</code></div>
    </form></section>
  </main>
}

function Shell({ user, setUser }) {
  const [page, navigate] = useHashPage(), features = new Set(user.features || [])
  const nav = [
    ['dashboard','Panel', features.has('DASH_CENTER_OVERVIEW')], ['users','Usuarios',features.has('IDENTITY_USER_VIEW')], ['roles','Roles',features.has('IDENTITY_ROLE_VIEW')],
    ['permissions','Permisos',features.has('IDENTITY_ROLE_VIEW')], ['audit','Auditoría',features.has('AUDIT_LOG_VIEW')], ['sessions','Mis sesiones',true]
  ].filter(x=>x[2])
  useEffect(() => { if (!nav.some(x => x[0] === page)) navigate(nav[0]?.[0] || 'sessions') }, [page, user.id])
  async function exit(){ await logout(); setUser(null) }
  return <div className="app-shell">
    <aside className="sidebar"><div className="side-brand"><div className="side-logo">I</div><div><strong>IAM Console</strong><small>Identity & Access</small></div></div>
      <nav>{nav.map(([key,label])=><button key={key} className={page===key?'active':''} onClick={()=>navigate(key)}><span>{icons[key]}</span>{label}</button>)}</nav>
      <div className="sidebar-user"><div className="avatar">{initials(user)}</div><div className="who"><b>{user.first_name} {user.last_name}</b><small>{user.roles?.join(', ')||'Sin rol'}</small></div><button title="Cerrar sesión" onClick={exit}>{icons.logout}</button></div>
    </aside>
    <main className="content"><Topbar page={page} user={user}/><div className="page-body">
      {page==='dashboard'&&<Dashboard/>}{page==='users'&&<Users canManage={features.has('IDENTITY_USER_MANAGE')} canAssign={features.has('IDENTITY_ROLE_ASSIGN')}/>} {page==='roles'&&<Roles/>}
      {page==='permissions'&&<Permissions/>}{page==='audit'&&<Audit/>}{page==='sessions'&&<Sessions user={user}/>} {!nav.some(x=>x[0]===page)&&page!=='permissions'&&page!=='sessions'&&<Dashboard/>}
    </div></main>
  </div>
}
function Topbar({page,user}) { const labels={dashboard:'Panel principal',users:'Gestión de usuarios',roles:'Roles del sistema',permissions:'Catálogo de permisos',audit:'Auditoría de accesos',sessions:'Sesiones activas'}; return <header className="topbar"><div><span className="eyebrow">Administración IAM</span><h2>{labels[page]||'Panel principal'}</h2></div><div className="top-profile"><span className="status-dot"/><div><b>{user.first_name}</b><small>Sesión activa</small></div></div></header> }

function Dashboard(){ const [d,setD]=useState(null),[error,setError]=useState(''); useEffect(()=>{api('/dashboard').then(setD).catch(e=>setError(e.message))},[]); if(error)return <ErrorBox text={error}/>; if(!d)return <Loading/>; const cards=[['Usuarios registrados',d.users,'Personas con identidad creada'],['Usuarios activos',d.active_users,'Cuentas habilitadas'],['Roles disponibles',d.roles,'Perfiles RBAC definidos'],['Sesiones activas',d.active_sessions,'Refresh tokens vigentes']]; return <><div className="welcome"><div><span className="eyebrow light">Resumen operativo</span><h1>Control de identidad y acceso</h1><p>Visibilidad rápida del estado de usuarios, roles y seguridad de autenticación.</p></div><div className="security-score"><b>{d.failed_logins_24h===0?'Sin alertas':'Revisar'}</b><span>{d.failed_logins_24h} intentos fallidos / 24 h</span></div></div><div className="stats-grid">{cards.map((c,i)=><div className="stat-card" key={c[0]}><div className="stat-icon">{['◉','✓','◆','◷'][i]}</div><span>{c[0]}</span><strong>{c[1]}</strong><small>{c[2]}</small></div>)}</div><div className="two-col"><section className="panel"><PanelTitle title="Arquitectura de seguridad" subtitle="Capas implementadas en el producto"/><div className="security-list">{[['Autenticación','JWT de corta duración + refresh token'],['Contraseñas','PBKDF2-SHA256 con salt aleatorio'],['Autorización','RBAC por feature y scope'],['Protección','Bloqueo tras 5 intentos fallidos'],['Trazabilidad','Auditoría de cada inicio de sesión']].map(x=><div key={x[0]}><span className="check">✓</span><div><b>{x[0]}</b><small>{x[1]}</small></div></div>)}</div></section><section className="panel"><PanelTitle title="Dominios de la base" subtitle="Esquemas PostgreSQL utilizados"/><div className="schema-grid">{['identity','rbac_catalog','rbac','session','identity_audit'].map(s=><span key={s}>{s}</span>)}</div></section></div></> }

function Users({canManage,canAssign}){ const [users,setUsers]=useState([]),[roles,setRoles]=useState([]),[q,setQ]=useState(''),[modal,setModal]=useState(null),[error,setError]=useState(''); async function load(){try{const [u,r]=await Promise.all([api(`/users?q=${encodeURIComponent(q)}`),api('/roles')]);setUsers(u);setRoles(r)}catch(e){setError(e.message)}} useEffect(()=>{load()},[]); async function openEdit(row){try{const full=await api(`/users/${row.ID}`);setModal({mode:'edit',user:{...row,...full}})}catch(e){setError(e.message)}} async function deactivate(id){if(!confirm('¿Desactivar este usuario y revocar sus sesiones?'))return;try{await api(`/users/${id}`,{method:'DELETE'});load()}catch(e){setError(e.message)}} return <section><div className="page-actions"><div className="search"><span>⌕</span><input value={q} onChange={e=>setQ(e.target.value)} onKeyDown={e=>e.key==='Enter'&&load()} placeholder="Buscar por nombre o correo..."/><button onClick={load}>Buscar</button></div>{canManage&&<button className="btn primary" onClick={()=>setModal({mode:'create'})}>+ Nuevo usuario</button>}</div>{error&&<ErrorBox text={error}/>}<div className="table-card"><table><thead><tr><th>Usuario</th><th>Tipo</th><th>Roles</th><th>Estado</th><th>Último acceso</th><th></th></tr></thead><tbody>{users.map(u=><tr key={u.ID}><td><div className="user-cell"><div className="avatar sm">{(u.FirstName?.[0]||'')+(u.LastName?.[0]||'')}</div><div><b>{u.FirstName} {u.LastName}</b><small>{u.Email}</small></div></div></td><td><span className="tag">{actorLabels[u.ActorType]||u.ActorType}</span></td><td>{u.Roles||<span className="muted">Sin rol</span>}</td><td><Status active={u.IsActive}/></td><td className="muted">{u.LastLoginAt?fmtDate(u.LastLoginAt):'Nunca'}</td><td className="row-actions">{canManage&&<button onClick={()=>openEdit(u)}>Editar</button>}{canManage&&u.IsActive&&<button className="danger-link" onClick={()=>deactivate(u.ID)}>Desactivar</button>}</td></tr>)}</tbody></table>{users.length===0&&<Empty text="No se encontraron usuarios."/>}</div>{modal&&<UserModal state={modal} roles={roles} canAssign={canAssign} onClose={()=>setModal(null)} onSaved={()=>{setModal(null);load()}}/>}</section> }

function UserModal({state,roles,canAssign,onClose,onSaved}) { const edit=state.mode==='edit', u=state.user||{}; const roleNameToId=Object.fromEntries(roles.map(r=>[r.Name,r.ID])); const [form,setForm]=useState({email:u.Email||u.email||'',password:'',first_name:u.FirstName||u.first_name||'',last_name:u.LastName||u.last_name||'',actor_type:u.ActorType||u.actor_type||'USER',is_active:u.IsActive??u.is_active??true,role_ids:(u.roles||[]).map(n=>roleNameToId[n]).filter(Boolean)}),[busy,setBusy]=useState(false),[error,setError]=useState(''); function change(k,v){setForm(x=>({...x,[k]:v}))} function toggleRole(id){change('role_ids',form.role_ids.includes(id)?form.role_ids.filter(x=>x!==id):[...form.role_ids,id])} async function save(e){e.preventDefault();setBusy(true);setError('');try{if(edit){await api(`/users/${u.ID||u.id}`,{method:'PUT',body:JSON.stringify(form)});if(canAssign)await api(`/users/${u.ID||u.id}/roles`,{method:'PUT',body:JSON.stringify({role_ids:form.role_ids})})}else await api('/users',{method:'POST',body:JSON.stringify(form)});onSaved()}catch(e){setError(e.message)}finally{setBusy(false)}} return <Modal onClose={onClose}><form onSubmit={save}><div className="modal-head"><div><span className="eyebrow">{edit?'Actualizar identidad':'Nueva identidad'}</span><h2>{edit?'Editar usuario':'Crear usuario'}</h2></div><button type="button" className="x" onClick={onClose}>×</button></div>{error&&<div className="alert error">{error}</div>}<div className="form-grid"><label>Nombre<input value={form.first_name} onChange={e=>change('first_name',e.target.value)} required/></label><label>Apellido<input value={form.last_name} onChange={e=>change('last_name',e.target.value)} required/></label><label className="span-2">Correo<input type="email" value={form.email} onChange={e=>change('email',e.target.value)} required/></label><label>Tipo de actor<select value={form.actor_type} onChange={e=>change('actor_type',e.target.value)}>{Object.entries(actorLabels).map(([k,v])=><option value={k} key={k}>{v}</option>)}</select></label><label>Contraseña<input type="password" minLength={edit?0:8} value={form.password} placeholder={edit?'Dejar vacía para conservar':'Mínimo 8 caracteres'} onChange={e=>change('password',e.target.value)} required={!edit}/></label>{edit&&<label className="toggle-line span-2"><input type="checkbox" checked={form.is_active} onChange={e=>change('is_active',e.target.checked)}/><span>Cuenta activa</span></label>}</div>{canAssign&&<div className="roles-picker"><b>Roles asignados</b><p className="muted">Selecciona uno o varios perfiles RBAC.</p><div>{roles.map(r=><label key={r.ID} className={form.role_ids.includes(r.ID)?'selected':''}><input type="checkbox" checked={form.role_ids.includes(r.ID)} onChange={()=>toggleRole(r.ID)}/><span><strong>{r.DisplayName}</strong><small>{r.Name}</small></span></label>)}</div></div>}<div className="modal-actions"><button type="button" className="btn secondary" onClick={onClose}>Cancelar</button><button className="btn primary" disabled={busy}>{busy?'Guardando...':'Guardar'}</button></div></form></Modal> }

function Roles(){const [roles,setRoles]=useState([]),[selected,setSelected]=useState(null),[error,setError]=useState('');useEffect(()=>{api('/roles').then(setRoles).catch(e=>setError(e.message))},[]);async function detail(id){try{setSelected(await api(`/roles/${id}`))}catch(e){setError(e.message)}}return <>{error&&<ErrorBox text={error}/>}<div className="role-grid">{roles.map(r=><button className="role-card" key={r.ID} onClick={()=>detail(r.ID)}><div className="role-icon">◆</div><div><span className="eyebrow">{r.Name}</span><h3>{r.DisplayName}</h3><p>{r.Description||'Rol del sistema IAM'}</p></div><div className="role-count"><b>{r.FeatureCount}</b><small>permisos</small></div></button>)}</div>{selected&&<Modal onClose={()=>setSelected(null)}><div className="modal-head"><div><span className="eyebrow">{selected.name}</span><h2>{selected.display_name}</h2></div><button className="x" onClick={()=>setSelected(null)}>×</button></div><p className="muted">{selected.description}</p><div className="permission-list">{groupBy(selected.features,'module').map(([module,fs])=><section key={module}><h4>{module}<span>{fs.length}</span></h4>{fs.map(f=><div key={f.code}><div><b>{f.name}</b><code>{f.code}</code></div><span className={`action ${f.action.toLowerCase()}`}>{f.action}</span><span className="scope">{f.scope}</span></div>)}</section>)}</div></Modal>}</>}

function Permissions(){const [mods,setMods]=useState([]),[error,setError]=useState('');useEffect(()=>{api('/catalog/modules').then(setMods).catch(e=>setError(e.message))},[]);return <>{error&&<ErrorBox text={error}/>}<div className="catalog">{mods.map(m=><details key={m.ID} open={m.Order<=2}><summary><div><span className="module-index">{String(m.Order).padStart(2,'0')}</span><div><b>{m.Name}</b><code>{m.Code}</code></div></div><span>{m.Features.length} features</span></summary><div className="feature-table">{m.Features.map(f=><div key={f.ID}><div><b>{f.Name}</b><code>{f.Code}</code></div><span className={`action ${f.Action.toLowerCase()}`}>{f.Action}</span></div>)}</div></details>)}</div></>}

function Audit(){const [logs,setLogs]=useState([]),[error,setError]=useState('');useEffect(()=>{api('/audit/logins?limit=200').then(setLogs).catch(e=>setError(e.message))},[]);return <>{error&&<ErrorBox text={error}/>}<div className="table-card"><table><thead><tr><th>Fecha</th><th>Correo</th><th>Resultado</th><th>IP</th><th>Navegador / dispositivo</th></tr></thead><tbody>{logs.map(x=><tr key={x.id}><td className="muted">{fmtDate(x.AttemptedAt)}</td><td><b>{x.Email}</b></td><td><span className={`outcome ${x.Outcome==='SUCCESS'?'ok':'bad'}`}>{outcomeLabels[x.Outcome]||x.Outcome}</span></td><td><code>{x.IPAddress||'—'}</code></td><td className="muted ua">{x.UserAgent||'—'}</td></tr>)}</tbody></table>{logs.length===0&&<Empty text="Aún no hay eventos de autenticación."/>}</div></>}

function Sessions({user}){const [rows,setRows]=useState([]),[error,setError]=useState('');useEffect(()=>{api('/sessions').then(setRows).catch(e=>setError(e.message))},[]);return <div className="two-col sessions-layout"><section className="panel profile-panel"><div className="big-avatar">{initials(user)}</div><h2>{user.first_name} {user.last_name}</h2><p>{user.email}</p><div className="profile-meta"><span>Tipo <b>{actorLabels[user.actor_type]||user.actor_type}</b></span><span>Roles <b>{user.roles?.join(', ')||'Sin rol'}</b></span><span>Permisos efectivos <b>{user.features?.length||0}</b></span></div></section><section className="panel"><PanelTitle title="Historial de sesiones" subtitle="Refresh tokens creados para tu usuario"/>{error&&<ErrorBox text={error}/>}<div className="session-list">{rows.map(x=><div key={x.id}><span className={`session-dot ${x.IsRevoked||new Date(x.ExpiresAt)<new Date()?'off':''}`}/><div><b>{x.DeviceHint||'Dispositivo desconocido'}</b><small>{x.IPAddress||'IP no registrada'} · creada {fmtDate(x.CreatedAt)}</small></div><span className="muted">{x.IsRevoked?'Revocada':new Date(x.ExpiresAt)<new Date()?'Expirada':'Activa'}</span></div>)}</div></section></div>}

function Modal({children,onClose}){useEffect(()=>{const f=e=>e.key==='Escape'&&onClose();addEventListener('keydown',f);return()=>removeEventListener('keydown',f)},[onClose]);return <div className="modal-backdrop" onMouseDown={e=>e.target===e.currentTarget&&onClose()}><div className="modal">{children}</div></div>}
function PanelTitle({title,subtitle}){return <div className="panel-title"><div><h3>{title}</h3><p>{subtitle}</p></div></div>}
function Loading(){return <div className="loading"><div className="spinner"/>Cargando datos...</div>}
function ErrorBox({text}){return <div className="alert error">{text}</div>}
function Empty({text}){return <div className="empty">{text}</div>}
function Status({active}){return <span className={`status ${active?'active':'inactive'}`}><i/>{active?'Activo':'Inactivo'}</span>}
function fmtDate(v){return new Intl.DateTimeFormat('es-CO',{dateStyle:'medium',timeStyle:'short'}).format(new Date(v))}
function initials(u){return `${u.first_name?.[0]||''}${u.last_name?.[0]||''}`.toUpperCase()}
function groupBy(items,key){const m=new Map();for(const i of items||[]){const k=i[key];if(!m.has(k))m.set(k,[]);m.get(k).push(i)}return [...m.entries()]}

export default App
