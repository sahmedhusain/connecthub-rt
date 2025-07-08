/**
 * Scroll Physics Animation System for Material Design 3
 * Provides smooth, physics-based scroll animations with performance optimization
 */

class ScrollPhysicsAnimator {
    constructor() {
        this.scrollElements = new Map();
        this.isScrolling = false;
        this.scrollDirection = 'down';
        this.lastScrollTop = 0;
        this.scrollVelocity = 0;
        this.animationFrame = null;
        this.intersectionObserver = null;
        
        this.init();
    }

    init() {
        this.setupIntersectionObserver();
        this.setupScrollListeners();
        this.scanForScrollElements();
        
        console.debug('[ScrollPhysics] Scroll physics animation system initialized');
    }

    /**
     * Setup intersection observer for scroll reveal animations
     */
    setupIntersectionObserver() {
        if (!('IntersectionObserver' in window)) return;

        this.intersectionObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.revealElement(entry.target);
                } else {
                    this.hideElement(entry.target);
                }
            });
        }, {
            rootMargin: '50px 0px -50px 0px',
            threshold: [0, 0.1, 0.5, 1]
        });
    }

    /**
     * Setup scroll event listeners with throttling
     */
    setupScrollListeners() {
        let ticking = false;

        const handleScroll = () => {
            if (!ticking) {
                requestAnimationFrame(() => {
                    this.updateScrollPhysics();
                    ticking = false;
                });
                ticking = true;
            }
        };

        window.addEventListener('scroll', handleScroll, { passive: true });
        
        // Handle scroll end
        let scrollTimeout;
        window.addEventListener('scroll', () => {
            this.isScrolling = true;
            clearTimeout(scrollTimeout);
            scrollTimeout = setTimeout(() => {
                this.isScrolling = false;
                this.resetScrollMomentum();
            }, 150);
        }, { passive: true });
    }

    /**
     * Update scroll physics calculations
     */
    updateScrollPhysics() {
        const currentScrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const scrollDelta = currentScrollTop - this.lastScrollTop;
        
        // Calculate scroll velocity
        this.scrollVelocity = scrollDelta;
        
        // Determine scroll direction
        if (scrollDelta > 0) {
            this.scrollDirection = 'down';
        } else if (scrollDelta < 0) {
            this.scrollDirection = 'up';
        }
        
        // Apply momentum effects
        this.applyScrollMomentum();
        
        // Apply parallax effects
        this.applyParallaxEffects(currentScrollTop);
        
        this.lastScrollTop = currentScrollTop;
    }

    /**
     * Apply scroll momentum to elements
     */
    applyScrollMomentum() {
        const momentumElements = document.querySelectorAll('.scroll-momentum');
        
        momentumElements.forEach(element => {
            if (this.isScrolling) {
                element.classList.remove('scrolling-up', 'scrolling-down');
                element.classList.add(`scrolling-${this.scrollDirection}`);
            }
        });
    }

    /**
     * Reset scroll momentum effects
     */
    resetScrollMomentum() {
        const momentumElements = document.querySelectorAll('.scroll-momentum');
        
        momentumElements.forEach(element => {
            element.classList.remove('scrolling-up', 'scrolling-down');
        });
    }

    /**
     * Apply parallax effects based on scroll position
     */
    applyParallaxEffects(scrollTop) {
        const parallaxElements = document.querySelectorAll('.scroll-parallax');
        
        parallaxElements.forEach(element => {
            const speed = parseFloat(element.dataset.parallaxSpeed) || 0.5;
            const yPos = -(scrollTop * speed);
            
            if (window.PerformanceOptimizer && window.PerformanceOptimizer.performanceLevel !== 'low') {
                element.style.transform = `translate3d(0, ${yPos}px, 0)`;
            }
        });
    }

    /**
     * Reveal element with animation
     */
    revealElement(element) {
        if (element.classList.contains('visible')) return;
        
        element.classList.add('visible');
        
        // Add stagger delay for elements in containers - ONLY for clickable containers
        const parentContainer = element.parentElement;
        if (parentContainer?.classList.contains('md3-stagger-container')) {
            // Check if container has clickable elements
            const hasClickableElements = parentContainer.querySelector('a, button, [role="button"], [tabindex="0"]');

            if (hasClickableElements) {
                const siblings = Array.from(parentContainer.children);
                const index = siblings.indexOf(element);
                const delay = index * 100; // 100ms stagger

                element.style.transitionDelay = `${delay}ms`;

                // Reset delay after animation
                setTimeout(() => {
                    element.style.transitionDelay = '';
                }, 800 + delay);
            }
        }
        
        // Trigger custom reveal event
        element.dispatchEvent(new CustomEvent('scrollReveal', {
            detail: { element, direction: this.scrollDirection }
        }));
    }

    /**
     * Hide element (optional - for elements that should hide when scrolled past)
     */
    hideElement(element) {
        if (element.dataset.scrollHide === 'true') {
            element.classList.remove('visible');
        }
    }

    /**
     * Scan for scroll elements and setup observers
     */
    scanForScrollElements() {
        // Scroll reveal elements
        const revealElements = document.querySelectorAll('.scroll-reveal');
        revealElements.forEach(element => {
            if (this.intersectionObserver) {
                this.intersectionObserver.observe(element);
            }
        });

        // Scroll scale elements
        const scaleElements = document.querySelectorAll('.scroll-scale');
        scaleElements.forEach(element => {
            this.setupScrollScaleElement(element);
        });

        // Scroll fade elements
        const fadeElements = document.querySelectorAll('.scroll-fade');
        fadeElements.forEach(element => {
            this.setupScrollFadeElement(element);
        });
    }

    /**
     * Setup scroll scale element
     */
    setupScrollScaleElement(element) {
        const trigger = parseFloat(element.dataset.scrollTrigger) || 0.5;
        
        if (this.intersectionObserver) {
            const scaleObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.intersectionRatio >= trigger) {
                        element.classList.add('scrolled');
                    } else {
                        element.classList.remove('scrolled');
                    }
                });
            }, { threshold: [0, trigger, 1] });
            
            scaleObserver.observe(element);
        }
    }

    /**
     * Setup scroll fade element
     */
    setupScrollFadeElement(element) {
        const fadeStart = parseFloat(element.dataset.fadeStart) || 0.8;
        
        if (this.intersectionObserver) {
            const fadeObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.intersectionRatio < fadeStart) {
                        element.classList.add('faded');
                    } else {
                        element.classList.remove('faded');
                    }
                });
            }, { threshold: [0, fadeStart, 1] });
            
            fadeObserver.observe(element);
        }
    }

    /**
     * Add scroll reveal to new elements
     */
    addScrollReveal(element, options = {}) {
        element.classList.add('scroll-reveal');
        
        if (options.delay) {
            element.style.transitionDelay = `${options.delay}ms`;
        }
        
        if (options.parallax) {
            element.classList.add('scroll-parallax');
            element.dataset.parallaxSpeed = options.parallaxSpeed || '0.5';
        }
        
        if (this.intersectionObserver) {
            this.intersectionObserver.observe(element);
        }
    }

    /**
     * Remove scroll animations from element
     */
    removeScrollAnimations(element) {
        element.classList.remove('scroll-reveal', 'scroll-parallax', 'scroll-scale', 'scroll-fade', 'visible', 'scrolled', 'faded');
        
        if (this.intersectionObserver) {
            this.intersectionObserver.unobserve(element);
        }
    }

    /**
     * Refresh scroll system (useful after DOM changes)
     */
    refresh() {
        this.scanForScrollElements();
    }

    /**
     * Destroy scroll physics system
     */
    destroy() {
        if (this.intersectionObserver) {
            this.intersectionObserver.disconnect();
        }
        
        if (this.animationFrame) {
            cancelAnimationFrame(this.animationFrame);
        }
        
        this.scrollElements.clear();
    }
}

// Initialize scroll physics system
const scrollPhysics = new ScrollPhysicsAnimator();

// Export for use in other modules
window.ScrollPhysics = scrollPhysics;

// Auto-setup scroll animations for new elements
const scrollMutationObserver = new MutationObserver((mutations) => {
    mutations.forEach(mutation => {
        mutation.addedNodes.forEach(node => {
            if (node.nodeType === Node.ELEMENT_NODE) {
                const scrollElements = node.querySelectorAll?.('.scroll-reveal, .scroll-scale, .scroll-fade, .scroll-parallax') || [];
                if (scrollElements.length > 0) {
                    scrollPhysics.refresh();
                }
            }
        });
    });
});

scrollMutationObserver.observe(document.body, {
    childList: true,
    subtree: true
});

console.debug('[ScrollPhysics] Scroll physics animation system loaded');
