// Глобальные переменные
let currentUser = null;
let currentFilter = '';
let currentFilterValue = '';

// Загрузка постов и категорий при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    fetchCurrentUser();
    loadPosts();
    loadCategories();
});

// Модальные окна
function showLogin() {
    renderLoginModal();
    document.getElementById('loginModal').style.display = 'block';
}
function showRegister() {
    renderRegisterModal();
    document.getElementById('registerModal').style.display = 'block';
}
function showCreatePost() {
    renderCreatePostModal();
    document.getElementById('createPostModal').style.display = 'block';
}
function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
}
window.onclick = function(event) {
    if (event.target.classList && event.target.classList.contains('modal')) {
        event.target.style.display = 'none';
    }
}

// Получить текущего пользователя
function fetchCurrentUser() {
    fetch('/api/user').then(r => r.json()).then(user => {
        if (user && user.id) {
            currentUser = user;
            renderAuthButtons();
            renderUserFilters();
        } else {
            currentUser = null;
            renderAuthButtons();
            renderUserFilters();
        }
    }).catch(() => {
        currentUser = null;
        renderAuthButtons();
        renderUserFilters();
    });
}

function renderAuthButtons() {
    const el = document.getElementById('auth-buttons');
    if (!el) return;
    if (currentUser) {
        el.innerHTML = `
            <span style="font-size:1.1rem;color:#1877f2;font-weight:500;margin-right:16px;">Привет, <b>${currentUser.username}</b>!</span>
            <button class="btn btn-primary" onclick="showCreatePost()">Создать пост</button>
            <button class="btn btn-secondary" onclick="logout()">Выйти</button>
        `;
    } else {
        el.innerHTML = `
            <button class="btn btn-primary" onclick="showLogin()">Войти</button>
            <button class="btn btn-secondary" onclick="showRegister()">Регистрация</button>
        `;
    }
}
function renderUserFilters() {
    const el = document.getElementById('user-filters');
    if (!el) return;
    if (currentUser) {
        el.innerHTML = `<h3>Мои фильтры</h3>
            <ul>
                <li><a href="#" onclick="loadPosts('created', '')">Мои посты</a></li>
                <li><a href="#" onclick="loadPosts('liked', '')">Понравившиеся</a></li>
            </ul>`;
    } else {
        el.innerHTML = '';
    }
}

// Загрузка постов
async function loadPosts(filter = '', value = '') {
    // Сохраняем текущий фильтр
    currentFilter = filter;
    currentFilterValue = value;
    const container = document.getElementById('posts-container');
    container.innerHTML = '<div class="loading">Загрузка постов...</div>';
    let url = '/api/posts';
    if (filter) {
        url += '?filter=' + filter;
        if (value) {
            url += '&value=' + encodeURIComponent(value);
        }
    }
    console.log('Loading posts from:', url);
    try {
        const response = await fetch(url);
        console.log('Posts response status:', response.status);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        const posts = await response.json();
        console.log('Posts loaded:', posts.length, 'posts');
        if (!posts || posts.length === 0) {
            container.innerHTML = '<p>Постов не найдено.</p>';
            return;
        }
        container.innerHTML = posts.map(post =>
            `<div class="post post-clickable" onclick="if(event.target === this || event.target.classList.contains('post-main')){loadPost(${post.id});}">
                ${renderAvatar(post.author_name)}
                <div class="post-main">
                    <div class="post-title">
                        <a href="/post/${post.id}" onclick="loadPost(${post.id}); return false;">${post.title}</a>
                        ${renderNewBadge(post.created)}
                    </div>
                    <div class="post-meta">
                        Автор: ${post.author_name} | ${new Date(post.created).toLocaleString('ru-RU')}
                    </div>
                    <div class="post-content">${post.content.substring(0, 200)}${post.content.length > 200 ? '...' : ''}</div>
                    <div class="post-categories">
                        ${post.categories ? post.categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : ''}
                    </div>
                    <div class="post-actions">
                        <button class="like-btn ${post.user_liked ? 'active' : ''}" onclick="toggleLike(${post.id}, null, true);event.stopPropagation();">👍 ${post.likes}</button>
                        <button class="dislike-btn ${post.user_disliked ? 'active' : ''}" onclick="toggleLike(${post.id}, null, false);event.stopPropagation();">👎 ${post.dislikes}</button>
                    </div>
                </div>
            </div>`
        ).join('');
    } catch (error) {
        console.error('Error loading posts:', error);
        container.innerHTML = '<p>Ошибка загрузки постов: ' + error.message + '</p>';
    }
}

// Загрузка категорий
async function loadCategories() {
    try {
        const response = await fetch('/api/categories');
        const categories = await response.json();
        const categoriesList = document.getElementById('categories-list');
        categories.forEach(category => {
            const li = document.createElement('li');
            li.innerHTML = `<a href="#" onclick="loadPosts('category', '${category.name}')">${category.name}</a>`;
            categoriesList.appendChild(li);
        });
    } catch (error) {
        console.error('Error loading categories:', error);
    }
}

// Загрузка конкретного поста
async function loadPost(postId) {
    const container = document.getElementById('posts-container');
    container.innerHTML = '<div class="loading">Загрузка поста...</div>';
    try {
        const response = await fetch('/api/post/' + postId);
        const data = await response.json();
        let commentForm = '';
        if (currentUser) {
            commentForm =
                `<form id="commentForm" style="margin-bottom: 20px;">
                    <div class="form-group">
                        <textarea name="content" placeholder="Написать комментарий..." required></textarea>
                    </div>
                    <input type="hidden" name="post_id" value="${data.post.id}">
                    <div id="commentError" class="error" style="color: red; margin-bottom: 10px;"></div>
                    <button type="submit" class="btn btn-primary">Добавить комментарий</button>
                </form>`;
        } else {
            commentForm = '<p>Войдите, чтобы оставить комментарий.</p>';
        }
        container.innerHTML =
            `<div class="post">
                ${renderAvatar(data.post.author_name)}
                <div class="post-main">
                    <div class="post-title">${data.post.title} ${renderNewBadge(data.post.created)}</div>
                    <div class="post-meta">Автор: ${data.post.author_name} | ${new Date(data.post.created).toLocaleString('ru-RU')}</div>
                    <div class="post-content">${data.post.content}</div>
                    <div class="post-categories">${(data.post.categories || []).map(cat => `<span class="category-tag">${cat}</span>`).join('')}</div>
                    <div class="post-actions">
                        <button class="like-btn ${data.post.user_liked ? 'active' : ''}" onclick="toggleLike(${data.post.id}, null, true)">👍 ${data.post.likes}</button>
                        <button class="dislike-btn ${data.post.user_disliked ? 'active' : ''}" onclick="toggleLike(${data.post.id}, null, false)">👎 ${data.post.dislikes}</button>
                    </div>
                </div>
            </div>
            <div style="margin-top: 30px;">
                <h3>Комментарии</h3>
                ${commentForm}
                <div id="comments-container">
                    ${(data.comments || []).map(comment =>
                        `<div class="post">
                            ${renderAvatar(comment.author_name)}
                            <div class="post-main">
                                <div class="post-meta">${comment.author_name} | ${new Date(comment.created).toLocaleString('ru-RU')}</div>
                                <div class="post-content">${comment.content}</div>
                                <div class="post-actions">
                                    <button class="like-btn ${comment.user_liked ? 'active' : ''}" onclick="toggleLike(null, ${comment.id}, true)">👍 ${comment.likes}</button>
                                    <button class="dislike-btn ${comment.user_disliked ? 'active' : ''}" onclick="toggleLike(null, ${comment.id}, false)">👎 ${comment.dislikes}</button>
                                </div>
                            </div>
                        </div>`
                    ).join('')}
                </div>
            </div>
            <button class="btn btn-secondary" onclick="loadPosts()" style="margin-top: 20px;">← Назад к постам</button>`;
        if (currentUser) {
            document.getElementById('commentForm').addEventListener('submit', handleCommentSubmit);
        }
    } catch (error) {
        container.innerHTML = '<p>Ошибка загрузки поста.</p>';
    }
}

// Лайк/дизлайк
async function toggleLike(postId, commentId, isLike) {
    if (!currentUser) {
        alert('Войдите, чтобы ставить лайки');
        return;
    }
    console.log('Toggle like called:', { postId, commentId, isLike });
    const formData = new FormData();
    if (postId) formData.append('post_id', postId);
    if (commentId) formData.append('comment_id', commentId);
    formData.append('is_like', isLike);
    const urlEncodedData = new URLSearchParams(formData);
    console.log('Sending data:', urlEncodedData.toString());
    try {
        const response = await fetch('/api/like', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: urlEncodedData
        });
        console.log('Response status:', response.status);
        if (response.ok) {
            const result = await response.json();
            console.log('Like result:', result);
            await fetchCurrentUser();
            // Если лайкаем пост или комментарий внутри просмотра поста, обновляем только этот пост
            if ((postId || commentId) && document.getElementById('posts-container').querySelector('.post-title') && document.getElementById('posts-container').innerHTML.includes('Назад к постам')) {
                // postId всегда есть для поста, для комментария берём скрытое поле post_id
                let pid = postId;
                if (!pid && commentId) {
                    // ищем скрытое поле post_id в форме комментария
                    const form = document.getElementById('commentForm');
                    if (form) {
                        pid = form.querySelector('input[name="post_id"]').value;
                    }
                }
                if (pid) {
                    loadPost(pid);
                    return;
                }
            }
            loadPosts(currentFilter, currentFilterValue);
        } else {
            const error = await response.json();
            console.error('Like error:', error);
        }
    } catch (error) {
        console.error('Error toggling like:', error);
    }
}

// Обработчики форм
function renderLoginModal() {
    document.getElementById('loginModal').innerHTML = `
        <div class="modal-content">
            <span class="close" onclick="closeModal('loginModal')">&times;</span>
            <h2>Вход</h2>
            <form id="loginForm">
                <div class="form-group">
                    <label for="loginEmail">Email:</label>
                    <input type="email" id="loginEmail" name="email" required>
                </div>
                <div class="form-group">
                    <label for="loginPassword">Пароль:</label>
                    <input type="password" id="loginPassword" name="password" required>
                </div>
                <div id="loginError" class="error"></div>
                <button type="submit" class="btn btn-primary">Войти</button>
            </form>
        </div>`;
    document.getElementById('loginForm').addEventListener('submit', async function(e) {
        e.preventDefault();
        const formData = new FormData(this);
        const urlEncodedData = new URLSearchParams(formData);
        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: urlEncodedData
            });
            if (response.ok) {
                closeModal('loginModal');
                location.reload();
            } else {
                const data = await response.json();
                document.getElementById('loginError').textContent = data.error || 'Ошибка входа';
            }
        } catch (error) {
            document.getElementById('loginError').textContent = 'Ошибка входа';
        }
    });
}
function renderRegisterModal() {
    document.getElementById('registerModal').innerHTML = `
        <div class="modal-content">
            <span class="close" onclick="closeModal('registerModal')">&times;</span>
            <h2>Регистрация</h2>
            <form id="registerForm">
                <div class="form-group">
                    <label for="registerUsername">Имя пользователя:</label>
                    <input type="text" id="registerUsername" name="username" required>
                </div>
                <div class="form-group">
                    <label for="registerEmail">Email:</label>
                    <input type="email" id="registerEmail" name="email" required>
                </div>
                <div class="form-group">
                    <label for="registerPassword">Пароль:</label>
                    <input type="password" id="registerPassword" name="password" required>
                </div>
                <div id="registerError" class="error"></div>
                <button type="submit" class="btn btn-primary">Зарегистрироваться</button>
            </form>
        </div>`;
    document.getElementById('registerForm').addEventListener('submit', async function(e) {
        e.preventDefault();
        const formData = new FormData(this);
        const urlEncodedData = new URLSearchParams(formData);
        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: urlEncodedData
            });
            if (response.ok) {
                closeModal('registerModal');
                alert('Регистрация успешна! Теперь войдите в систему.');
            } else {
                const data = await response.json();
                document.getElementById('registerError').textContent = data.error || 'Ошибка регистрации';
            }
        } catch (error) {
            document.getElementById('registerError').textContent = 'Ошибка регистрации';
        }
    });
}
function renderCreatePostModal() {
    // Сначала загрузим категории для выпадающего списка
    fetch('/api/categories')
        .then(response => response.json())
        .then(categories => {
            const categoryOptions = categories.map(cat => 
                `<option value="${cat.name}">${cat.name}</option>`
            ).join('');
            
            document.getElementById('createPostModal').innerHTML = `
                <div class="modal-content">
                    <span class="close" onclick="closeModal('createPostModal')">&times;</span>
                    <h2>Создать пост</h2>
                    <form id="createPostForm">
                        <div class="form-group">
                            <label for="postTitle">Заголовок (5-100 символов):</label>
                            <input type="text" id="postTitle" name="title" required minlength="5" maxlength="100">
                        </div>
                        <div class="form-group">
                            <label for="postContent">Содержание (10-2000 символов):</label>
                            <textarea id="postContent" name="content" required minlength="10" maxlength="2000"></textarea>
                        </div>
                        <div class="form-group">
                            <label for="postCategories">Категории (выберите до 4):</label>
                            <select id="postCategories" name="categories" multiple size="5">
                                ${categoryOptions}
                            </select>
                            <small>Удерживайте Ctrl (Cmd на Mac) для выбора нескольких категорий</small>
                        </div>
                        <div id="createPostError" class="error"></div>
                        <button type="submit" class="btn btn-primary">Создать</button>
                    </form>
                </div>`;
            
            document.getElementById('createPostForm').addEventListener('submit', async function(e) {
                e.preventDefault();
                const formData = new FormData(this);
                
                // Получить выбранные категории из select
                const categorySelect = document.getElementById('postCategories');
                const selectedCategories = Array.from(categorySelect.selectedOptions).map(option => option.value);
                formData.set('categories', selectedCategories.join(','));
                
                const urlEncodedData = new URLSearchParams(formData);
                try {
                    const response = await fetch('/api/posts', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded',
                        },
                        body: urlEncodedData
                    });
                    if (response.ok) {
                        closeModal('createPostModal');
                        this.reset();
                        loadPosts();
                    } else {
                        const data = await response.json();
                        document.getElementById('createPostError').textContent = data.error || 'Ошибка создания поста';
                    }
                } catch (error) {
                    document.getElementById('createPostError').textContent = 'Ошибка создания поста';
                }
            });
        })
        .catch(error => {
            console.error('Error loading categories:', error);
            // Fallback без категорий
            document.getElementById('createPostModal').innerHTML = `
                <div class="modal-content">
                    <span class="close" onclick="closeModal('createPostModal')">&times;</span>
                    <h2>Создать пост</h2>
                    <form id="createPostForm">
                        <div class="form-group">
                            <label for="postTitle">Заголовок (5-100 символов):</label>
                            <input type="text" id="postTitle" name="title" required minlength="5" maxlength="100">
                        </div>
                        <div class="form-group">
                            <label for="postContent">Содержание (10-2000 символов):</label>
                            <textarea id="postContent" name="content" required minlength="10" maxlength="2000"></textarea>
                        </div>
                        <div id="createPostError" class="error"></div>
                        <button type="submit" class="btn btn-primary">Создать</button>
                    </form>
                </div>`;
        });
}
async function handleCommentSubmit(e) {
    e.preventDefault();
    const formData = new FormData(this);
    const urlEncodedData = new URLSearchParams(formData);
    
    // Clear previous error
    const errorElement = document.getElementById('commentError');
    if (errorElement) {
        errorElement.textContent = '';
    }
    
    try {
        const response = await fetch('/api/comments', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: urlEncodedData
        });
        if (response.ok) {
            this.reset();
            const postId = formData.get('post_id');
            loadPost(postId);
        } else {
            const data = await response.json();
            const errorMessage = data.error || 'Ошибка создания комментария';
            
            // Show error message
            if (errorElement) {
                errorElement.textContent = errorMessage;
            } else {
                // If no error element exists, create one or show alert
                alert(errorMessage);
            }
        }
    } catch (error) {
        console.error('Error creating comment:', error);
        const errorMessage = 'Ошибка создания комментария';
        if (errorElement) {
            errorElement.textContent = errorMessage;
        } else {
            alert(errorMessage);
        }
    }
}
async function logout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST'
        });
        if (response.ok) {
            location.reload();
        }
    } catch (error) {
        console.error('Error logging out:', error);
    }
}

// Вспомогательная функция для аватарки
function renderAvatar(username) {
    const initial = username && username.length > 0 ? username[0].toUpperCase() : '?';
    return `<div class="avatar">${initial}</div>`;
}
// Вспомогательная функция для бейджа NEW
function renderNewBadge(created) {
    const createdDate = new Date(created);
    const now = new Date();
    const diff = (now - createdDate) / (1000 * 60 * 60 * 24);
    if (diff < 1) {
        return '<span class="badge-new">NEW</span>';
    }
    return '';
} 