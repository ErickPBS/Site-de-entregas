/* ===== FAST DELIVERY - MAIN JAVASCRIPT ===== */

// Global Variables
const FastDelivery = {
    // Configuration
    config: {
        animationDuration: 300,
        scrollOffset: 100,
        apiUrl: '/api',
        version: '1.0.0'
    },
    
    // State Management
    state: {
        isLoading: false,
        currentUser: null,
        cart: [],
        notifications: []
    },
    
    // Utility Functions
    utils: {
        // Debounce function
        debounce: function(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        },
        
        // Throttle function
        throttle: function(func, limit) {
            let inThrottle;
            return function() {
                const args = arguments;
                const context = this;
                if (!inThrottle) {
                    func.apply(context, args);
                    inThrottle = true;
                    setTimeout(() => inThrottle = false, limit);
                }
            }
        },
        
        // Format currency
        formatCurrency: function(value) {
            return new Intl.NumberFormat('pt-BR', {
                style: 'currency',
                currency: 'BRL'
            }).format(value);
        },
        
        // Format date
        formatDate: function(date) {
            return new Intl.DateTimeFormat('pt-BR', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            }).format(new Date(date));
        },
        
        // Generate unique ID
        generateId: function() {
            return Date.now().toString(36) + Math.random().toString(36).substr(2);
        },
        
        // Validate email
        validateEmail: function(email) {
            const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return re.test(email);
        },
        
        // Validate phone
        validatePhone: function(phone) {
            const re = /^\(\d{2}\)\s\d{4,5}-\d{4}$/;
            return re.test(phone);
        },
        
        // Format phone
        formatPhone: function(phone) {
            const cleaned = phone.replace(/\D/g, '');
            const match = cleaned.match(/^(\d{2})(\d{4,5})(\d{4})$/);
            if (match) {
                return `(${match[1]}) ${match[2]}-${match[3]}`;
            }
            return phone;
        },
        
        // Get query parameter
        getQueryParam: function(param) {
            const urlParams = new URLSearchParams(window.location.search);
            return urlParams.get(param);
        },
        
        // Set query parameter
        setQueryParam: function(param, value) {
            const url = new URL(window.location);
            url.searchParams.set(param, value);
            window.history.pushState({}, '', url);
        }
    },
    
    // Animation Functions
    animations: {
        // Fade in animation
        fadeIn: function(element, duration = 300) {
            element.style.opacity = '0';
            element.style.display = 'block';
            
            let start = null;
            function animate(timestamp) {
                if (!start) start = timestamp;
                const progress = timestamp - start;
                const opacity = Math.min(progress / duration, 1);
                
                element.style.opacity = opacity;
                
                if (progress < duration) {
                    requestAnimationFrame(animate);
                }
            }
            requestAnimationFrame(animate);
        },
        
        // Fade out animation
        fadeOut: function(element, duration = 300) {
            let start = null;
            function animate(timestamp) {
                if (!start) start = timestamp;
                const progress = timestamp - start;
                const opacity = Math.max(1 - (progress / duration), 0);
                
                element.style.opacity = opacity;
                
                if (progress < duration) {
                    requestAnimationFrame(animate);
                } else {
                    element.style.display = 'none';
                }
            }
            requestAnimationFrame(animate);
        },
        
        // Slide up animation
        slideUp: function(element, duration = 300) {
            element.style.height = element.scrollHeight + 'px';
            element.style.overflow = 'hidden';
            
            let start = null;
            function animate(timestamp) {
                if (!start) start = timestamp;
                const progress = timestamp - start;
                const height = Math.max(element.scrollHeight * (1 - progress / duration), 0);
                
                element.style.height = height + 'px';
                
                if (progress < duration) {
                    requestAnimationFrame(animate);
                } else {
                    element.style.display = 'none';
                    element.style.height = '';
                    element.style.overflow = '';
                }
            }
            requestAnimationFrame(animate);
        },
        
        // Slide down animation
        slideDown: function(element, duration = 300) {
            element.style.display = 'block';
            const height = element.scrollHeight;
            element.style.height = '0px';
            element.style.overflow = 'hidden';
            
            let start = null;
            function animate(timestamp) {
                if (!start) start = timestamp;
                const progress = timestamp - start;
                const currentHeight = Math.min(height * (progress / duration), height);
                
                element.style.height = currentHeight + 'px';
                
                if (progress < duration) {
                    requestAnimationFrame(animate);
                } else {
                    element.style.height = '';
                    element.style.overflow = '';
                }
            }
            requestAnimationFrame(animate);
        }
    },
    
    // UI Components
    ui: {
        // Show loading state
        showLoading: function(element) {
            if (element) {
                element.classList.add('loading');
                element.disabled = true;
            }
            FastDelivery.state.isLoading = true;
        },
        
        // Hide loading state
        hideLoading: function(element) {
            if (element) {
                element.classList.remove('loading');
                element.disabled = false;
            }
            FastDelivery.state.isLoading = false;
        },
        
        // Show notification
        showNotification: function(message, type = 'info', duration = 5000) {
            const notification = document.createElement('div');
            notification.className = `notification notification-${type} animate-slide-down`;
            notification.innerHTML = `
                <div class="notification-content">
                    <i class="fas fa-${this.getNotificationIcon(type)}"></i>
                    <span>${message}</span>
                </div>
                <button class="notification-close">
                    <i class="fas fa-times"></i>
                </button>
            `;
            
            // Add to container
            let container = document.querySelector('.notifications-container');
            if (!container) {
                container = document.createElement('div');
                container.className = 'notifications-container';
                document.body.appendChild(container);
            }
            
            container.appendChild(notification);
            
            // Auto remove
            setTimeout(() => {
                this.hideNotification(notification);
            }, duration);
            
            // Close button
            notification.querySelector('.notification-close').addEventListener('click', () => {
                this.hideNotification(notification);
            });
            
            return notification;
        },
        
        // Hide notification
        hideNotification: function(notification) {
            notification.classList.add('animate-fade-out');
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        },
        
        // Get notification icon
        getNotificationIcon: function(type) {
            const icons = {
                success: 'check-circle',
                error: 'exclamation-circle',
                warning: 'exclamation-triangle',
                info: 'info-circle'
            };
            return icons[type] || 'info-circle';
        },
        
        // Show modal
        showModal: function(content, options = {}) {
            const modal = document.createElement('div');
            modal.className = 'modal-overlay';
            modal.innerHTML = `
                <div class="modal animate-scale-in">
                    <div class="modal-header">
                        <h3>${options.title || 'Modal'}</h3>
                        <button class="modal-close">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="modal-body">
                        ${content}
                    </div>
                    ${options.footer ? `<div class="modal-footer">${options.footer}</div>` : ''}
                </div>
            `;
            
            document.body.appendChild(modal);
            document.body.style.overflow = 'hidden';
            
            // Close events
            modal.querySelector('.modal-close').addEventListener('click', () => {
                this.hideModal(modal);
            });
            
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideModal(modal);
                }
            });
            
            return modal;
        },
        
        // Hide modal
        hideModal: function(modal) {
            modal.querySelector('.modal').classList.add('animate-fade-out');
            setTimeout(() => {
                if (modal.parentNode) {
                    modal.parentNode.removeChild(modal);
                    document.body.style.overflow = '';
                }
            }, 300);
        },
        
        // Show tooltip
        showTooltip: function(element, text, position = 'top') {
            const tooltip = document.createElement('div');
            tooltip.className = `tooltip tooltip-${position}`;
            tooltip.textContent = text;
            
            document.body.appendChild(tooltip);
            
            const rect = element.getBoundingClientRect();
            const tooltipRect = tooltip.getBoundingClientRect();
            
            let top, left;
            
            switch (position) {
                case 'top':
                    top = rect.top - tooltipRect.height - 10;
                    left = rect.left + (rect.width - tooltipRect.width) / 2;
                    break;
                case 'bottom':
                    top = rect.bottom + 10;
                    left = rect.left + (rect.width - tooltipRect.width) / 2;
                    break;
                case 'left':
                    top = rect.top + (rect.height - tooltipRect.height) / 2;
                    left = rect.left - tooltipRect.width - 10;
                    break;
                case 'right':
                    top = rect.top + (rect.height - tooltipRect.height) / 2;
                    left = rect.right + 10;
                    break;
            }
            
            tooltip.style.top = top + 'px';
            tooltip.style.left = left + 'px';
            tooltip.classList.add('animate-fade-in');
            
            return tooltip;
        }
    },
    
    // Form Handling
    forms: {
        // Validate form
        validate: function(form) {
            const errors = [];
            const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
            
            inputs.forEach(input => {
                if (!input.value.trim()) {
                    errors.push(`${input.name || input.id} é obrigatório`);
                    this.showFieldError(input, 'Campo obrigatório');
                } else {
                    this.clearFieldError(input);
                    
                    // Specific validations
                    if (input.type === 'email' && !FastDelivery.utils.validateEmail(input.value)) {
                        errors.push('E-mail inválido');
                        this.showFieldError(input, 'E-mail inválido');
                    }
                    
                    if (input.type === 'tel' && !FastDelivery.utils.validatePhone(input.value)) {
                        errors.push('Telefone inválido');
                        this.showFieldError(input, 'Telefone inválido');
                    }
                }
            });
            
            return errors;
        },
        
        // Show field error
        showFieldError: function(field, message) {
            this.clearFieldError(field);
            
            const error = document.createElement('div');
            error.className = 'form-error';
            error.textContent = message;
            
            field.parentNode.appendChild(error);
            field.classList.add('error');
        },
        
        // Clear field error
        clearFieldError: function(field) {
            const error = field.parentNode.querySelector('.form-error');
            if (error) {
                error.remove();
            }
            field.classList.remove('error');
        },
        
        // Serialize form data
        serialize: function(form) {
            const formData = new FormData(form);
            const data = {};
            
            for (let [key, value] of formData.entries()) {
                if (data[key]) {
                    if (Array.isArray(data[key])) {
                        data[key].push(value);
                    } else {
                        data[key] = [data[key], value];
                    }
                } else {
                    data[key] = value;
                }
            }
            
            return data;
        }
    },
    
    // API Functions
    api: {
        // Make API request
        request: async function(url, options = {}) {
            const defaultOptions = {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest'
                }
            };
            
            const config = { ...defaultOptions, ...options };
            
            try {
                const response = await fetch(FastDelivery.config.apiUrl + url, config);
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.message || 'Erro na requisição');
                }
                
                return data;
            } catch (error) {
                console.error('API Error:', error);
                throw error;
            }
        },
        
        // GET request
        get: function(url) {
            return this.request(url);
        },
        
        // POST request
        post: function(url, data) {
            return this.request(url, {
                method: 'POST',
                body: JSON.stringify(data)
            });
        },
        
        // PUT request
        put: function(url, data) {
            return this.request(url, {
                method: 'PUT',
                body: JSON.stringify(data)
            });
        },
        
        // DELETE request
        delete: function(url) {
            return this.request(url, {
                method: 'DELETE'
            });
        }
    },
    
    // Storage Functions
    storage: {
        // Set item in localStorage
        set: function(key, value) {
            try {
                localStorage.setItem(key, JSON.stringify(value));
            } catch (error) {
                console.error('Storage Error:', error);
            }
        },
        
        // Get item from localStorage
        get: function(key, defaultValue = null) {
            try {
                const item = localStorage.getItem(key);
                return item ? JSON.parse(item) : defaultValue;
            } catch (error) {
                console.error('Storage Error:', error);
                return defaultValue;
            }
        },
        
        // Remove item from localStorage
        remove: function(key) {
            try {
                localStorage.removeItem(key);
            } catch (error) {
                console.error('Storage Error:', error);
            }
        },
        
        // Clear all localStorage
        clear: function() {
            try {
                localStorage.clear();
            } catch (error) {
                console.error('Storage Error:', error);
            }
        }
    }
};

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize header scroll effect
    const header = document.querySelector('.header');
    if (header) {
        window.addEventListener('scroll', FastDelivery.utils.throttle(() => {
            if (window.scrollY > FastDelivery.config.scrollOffset) {
                header.style.background = 'rgba(255, 255, 255, 0.98)';
                header.style.boxShadow = '0 5px 30px rgba(0, 0, 0, 0.1)';
            } else {
                header.style.background = 'rgba(255, 255, 255, 0.95)';
                header.style.boxShadow = '0 2px 20px rgba(0, 0, 0, 0.1)';
            }
        }, 100));
    }
    
    // Initialize smooth scrolling
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
    
    // Initialize mobile menu
    const mobileToggle = document.querySelector('.mobile-menu-toggle');
    const mobileMenu = document.querySelector('.mobile-menu');
    const mobileClose = document.querySelector('.mobile-menu-close');
    
    if (mobileToggle && mobileMenu) {
        mobileToggle.addEventListener('click', () => {
            mobileMenu.classList.add('active');
        });
    }
    
    if (mobileClose && mobileMenu) {
        mobileClose.addEventListener('click', () => {
            mobileMenu.classList.remove('active');
        });
    }
    
    // Initialize intersection observer for animations
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-fade-in-up');
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);
    
    // Observe elements with animation classes
    document.querySelectorAll('.card, .feature-card, .animate-on-scroll').forEach(el => {
        observer.observe(el);
    });
    
    // Initialize form enhancements
    document.querySelectorAll('input[type="tel"]').forEach(input => {
        input.addEventListener('input', function() {
            this.value = FastDelivery.utils.formatPhone(this.value);
        });
    });
    
    // Initialize tooltips
    document.querySelectorAll('[data-tooltip]').forEach(element => {
        element.addEventListener('mouseenter', function() {
            const tooltip = FastDelivery.ui.showTooltip(
                this,
                this.dataset.tooltip,
                this.dataset.tooltipPosition || 'top'
            );
            this._tooltip = tooltip;
        });
        
        element.addEventListener('mouseleave', function() {
            if (this._tooltip) {
                this._tooltip.remove();
                this._tooltip = null;
            }
        });
    });
    
    // Initialize loading states
    document.querySelectorAll('form').forEach(form => {
        form.addEventListener('submit', function(e) {
            const submitBtn = this.querySelector('button[type="submit"]');
            if (submitBtn) {
                FastDelivery.ui.showLoading(submitBtn);
            }
        });
    });
    
    // Initialize click animations
    document.querySelectorAll('.btn, .card').forEach(element => {
        element.addEventListener('click', function(e) {
            const ripple = document.createElement('span');
            ripple.className = 'ripple';
            
            const rect = this.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;
            
            ripple.style.width = ripple.style.height = size + 'px';
            ripple.style.left = x + 'px';
            ripple.style.top = y + 'px';
            
            this.appendChild(ripple);
            
            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });
    
    console.log('Fast Delivery initialized successfully!');
});

// Export for global use
window.FastDelivery = FastDelivery;
