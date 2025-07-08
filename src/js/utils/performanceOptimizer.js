/**
 * Performance Optimizer for Material Design 3 Animations
 * Manages animation performance, reduces motion, and optimizes for 60fps
 */

class PerformanceOptimizer {
    constructor() {
        this.isReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        this.isLowPowerMode = this.detectLowPowerMode();
        this.performanceLevel = this.detectPerformanceLevel();
        this.animationFrameId = null;
        this.observedElements = new Set();
        
        this.init();
    }

    init() {
        this.setupReducedMotionListener();
        this.setupIntersectionObserver();
        this.optimizeExistingAnimations();
        this.setupPerformanceMonitoring();
        
        console.debug('[PerformanceOptimizer] Initialized with performance level:', this.performanceLevel);
    }

    /**
     * Detect device performance level
     */
    detectPerformanceLevel() {
        const hardwareConcurrency = navigator.hardwareConcurrency || 2;
        const memory = navigator.deviceMemory || 2;
        
        if (hardwareConcurrency >= 8 && memory >= 8) {
            return 'high';
        } else if (hardwareConcurrency >= 4 && memory >= 4) {
            return 'medium';
        } else {
            return 'low';
        }
    }

    /**
     * Detect low power mode or battery saving
     */
    detectLowPowerMode() {
        // Check for battery API
        if ('getBattery' in navigator) {
            navigator.getBattery().then(battery => {
                return battery.level < 0.2 || battery.charging === false;
            });
        }
        return false;
    }

    /**
     * Setup reduced motion preference listener
     */
    setupReducedMotionListener() {
        const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
        mediaQuery.addEventListener('change', (e) => {
            this.isReducedMotion = e.matches;
            this.updateAnimationSettings();
        });
    }

    /**
     * Setup intersection observer for lazy animations
     */
    setupIntersectionObserver() {
        if (!('IntersectionObserver' in window)) return;

        this.intersectionObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.activateElement(entry.target);
                } else {
                    this.deactivateElement(entry.target);
                }
            });
        }, {
            rootMargin: '50px',
            threshold: 0.1
        });
    }

    /**
     * Optimize animations based on performance level
     */
    optimizeExistingAnimations() {
        const elements = document.querySelectorAll('.md3-enhanced, .post-card, .btn, .dropdown-menu');
        
        elements.forEach(element => {
            this.optimizeElement(element);
        });
    }

    /**
     * Optimize individual element
     */
    optimizeElement(element) {
        if (this.isReducedMotion || this.performanceLevel === 'low') {
            element.classList.add('reduced-motion');
            element.style.animationDuration = '0.01ms';
            element.style.transitionDuration = '0.01ms';
        } else if (this.performanceLevel === 'medium') {
            element.classList.add('medium-performance');
            this.applyMediumPerformanceOptimizations(element);
        } else {
            element.classList.add('high-performance');
            this.applyHighPerformanceOptimizations(element);
        }
    }

    /**
     * Apply medium performance optimizations
     */
    applyMediumPerformanceOptimizations(element) {
        element.style.willChange = 'transform, opacity';
        element.style.transform = 'translateZ(0)';
        element.style.backfaceVisibility = 'hidden';
        
        // Reduce animation complexity
        if (element.classList.contains('btn')) {
            element.style.animationDuration = '200ms';
        }
    }

    /**
     * Apply high performance optimizations
     */
    applyHighPerformanceOptimizations(element) {
        element.style.willChange = 'transform, opacity';
        element.style.transform = 'translate3d(0, 0, 0)';
        element.style.backfaceVisibility = 'hidden';
        element.style.perspective = '1000px';
        
        // Enable full animation complexity
        element.classList.add('gpu-accelerated');
    }

    /**
     * Activate element when in viewport
     */
    activateElement(element) {
        if (!this.isReducedMotion) {
            element.style.willChange = 'transform, opacity';
            element.classList.add('visible');
        }
        this.observedElements.add(element);
    }

    /**
     * Deactivate element when out of viewport
     */
    deactivateElement(element) {
        // Reset will-change to auto for performance
        setTimeout(() => {
            if (!element.matches(':hover, :focus, :active')) {
                element.style.willChange = 'auto';
            }
        }, 300);
        this.observedElements.delete(element);
    }

    /**
     * Setup performance monitoring
     */
    setupPerformanceMonitoring() {
        if ('PerformanceObserver' in window) {
            const observer = new PerformanceObserver((list) => {
                const entries = list.getEntries();
                entries.forEach(entry => {
                    if (entry.duration > 16.67) { // More than 60fps
                        console.warn('[PerformanceOptimizer] Slow animation detected:', entry);
                        this.handleSlowAnimation(entry);
                    }
                });
            });
            
            observer.observe({ entryTypes: ['measure', 'navigation'] });
        }
    }

    /**
     * Handle slow animations
     */
    handleSlowAnimation(entry) {
        // Reduce animation complexity if performance is poor
        if (this.performanceLevel !== 'low') {
            this.performanceLevel = 'medium';
            this.optimizeExistingAnimations();
        }
    }

    /**
     * Update animation settings based on preferences
     */
    updateAnimationSettings() {
        const root = document.documentElement;
        
        if (this.isReducedMotion) {
            root.style.setProperty('--transition-duration', '0.01ms');
            root.style.setProperty('--transition-duration-fast', '0.01ms');
            root.style.setProperty('--animation-duration', '0.01ms');
        } else {
            root.style.removeProperty('--transition-duration');
            root.style.removeProperty('--transition-duration-fast');
            root.style.removeProperty('--animation-duration');
        }
    }

    /**
     * Observe element for lazy animation
     */
    observeElement(element) {
        if (this.intersectionObserver) {
            element.classList.add('lazy-animate');
            this.intersectionObserver.observe(element);
        }
    }

    /**
     * Unobserve element
     */
    unobserveElement(element) {
        if (this.intersectionObserver) {
            this.intersectionObserver.unobserve(element);
        }
    }

    /**
     * Force GPU acceleration for critical animations
     */
    forceGPUAcceleration(element) {
        element.style.transform = 'translate3d(0, 0, 0)';
        element.style.backfaceVisibility = 'hidden';
        element.style.perspective = '1000px';
        element.style.willChange = 'transform, opacity';
    }

    /**
     * Reset GPU acceleration
     */
    resetGPUAcceleration(element) {
        element.style.willChange = 'auto';
        element.style.transform = '';
        element.style.backfaceVisibility = '';
        element.style.perspective = '';
    }

    /**
     * Get current performance metrics
     */
    getPerformanceMetrics() {
        return {
            performanceLevel: this.performanceLevel,
            isReducedMotion: this.isReducedMotion,
            isLowPowerMode: this.isLowPowerMode,
            observedElements: this.observedElements.size,
            hardwareConcurrency: navigator.hardwareConcurrency || 'unknown',
            deviceMemory: navigator.deviceMemory || 'unknown'
        };
    }

    /**
     * Cleanup
     */
    destroy() {
        if (this.intersectionObserver) {
            this.intersectionObserver.disconnect();
        }
        if (this.animationFrameId) {
            cancelAnimationFrame(this.animationFrameId);
        }
        this.observedElements.clear();
    }
}

// Initialize performance optimizer
const performanceOptimizer = new PerformanceOptimizer();

// Export for use in other modules
window.PerformanceOptimizer = performanceOptimizer;

// Auto-optimize new elements
const autoOptimize = new MutationObserver((mutations) => {
    mutations.forEach(mutation => {
        mutation.addedNodes.forEach(node => {
            if (node.nodeType === Node.ELEMENT_NODE) {
                const elements = node.querySelectorAll?.('.md3-enhanced, .post-card, .btn, .dropdown-menu') || [];
                elements.forEach(element => {
                    performanceOptimizer.optimizeElement(element);
                    performanceOptimizer.observeElement(element);
                });
            }
        });
    });
});

autoOptimize.observe(document.body, {
    childList: true,
    subtree: true
});

console.debug('[PerformanceOptimizer] Performance optimization system loaded');
