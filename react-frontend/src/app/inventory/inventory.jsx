import React, { useState, useEffect, useCallback } from 'react';
import { Trash2, PlusCircle, Loader2, Pencil, Save, X, AlertTriangle } from 'lucide-react';

// --- SHADCN/UI IMPORTS ---
// NOTE: These imports are correct for a real Next.js project with shadcn/ui installed.
// They replace the local component definitions from the previous version.
// This file assumes the following paths exist in your project:
// - @/components/ui/button.jsx
// - @/components/ui/input.jsx
// - @/components/ui/card.jsx
// - @/components/ui/alert.jsx
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';


// Base URL for the Go API (must match main.go)
const API_BASE_URL = 'http://localhost:8080';


// --- API Functions (Unchanged) ---

const api = {
    fetchProducts: async () => {
        const response = await fetch(`${API_BASE_URL}/products`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        return response.json();
    },
    createProduct: async (productData) => {
        const response = await fetch(`${API_BASE_URL}/products`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(productData),
        });
        if (!response.ok) throw new Error("Failed to create product on server.");
        return response.json();
    },
    updateProduct: async (id, productData) => {
        const response = await fetch(`${API_BASE_URL}/products/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(productData),
        });
        if (!response.ok) throw new Error("Failed to update product on server.");
        return response.json();
    },
    deleteProduct: async (id) => {
        const response = await fetch(`${API_BASE_URL}/products/${id}`, {
            method: 'DELETE',
        });
        if (!response.ok) throw new Error("Failed to delete product.");
        return response.json();
    }
};

// --- Child Components ---

const ProductForm = ({ onProductAdded, onError }) => {
    const [newProduct, setNewProduct] = useState({ name: '', price: '' });
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setNewProduct(prev => ({ ...prev, [name]: value }));
    };

    const handleAddProduct = async (e) => {
        e.preventDefault();
        setIsSubmitting(true);
        onError(null);

        const dataToSend = {
            name: newProduct.name.trim(),
            price: parseFloat(newProduct.price), 
        };
        
        if (!dataToSend.name || isNaN(dataToSend.price) || dataToSend.price <= 0) {
            onError("Please enter a valid product name and a price greater than zero.");
            setIsSubmitting(false);
            return;
        }

        try {
            await api.createProduct(dataToSend);
            setNewProduct({ name: '', price: '' });
            onProductAdded();
        } catch (e) {
            console.error("Error creating product:", e);
            onError("Could not add product. Check Go server logs.");
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Card className="h-fit p-6"> {/* Ensure padding is applied to Card */}
            <h2 className="text-2xl font-bold text-gray-800 mb-6 flex items-center">
                <PlusCircle className="w-6 h-6 mr-3 text-indigo-600" />
                Add New Product
            </h2>
            <form onSubmit={handleAddProduct} className="space-y-4">
                <div className="space-y-1">
                    <label htmlFor="name" className="text-sm font-medium text-gray-700">Product Name</label>
                    <Input
                        type="text"
                        id="name"
                        name="name"
                        value={newProduct.name}
                        onChange={handleInputChange}
                        disabled={isSubmitting}
                        placeholder="e.g., Ultra 4K Monitor"
                        required
                    />
                </div>
                <div className="space-y-1">
                    <label htmlFor="price" className="text-sm font-medium text-gray-700">Price (USD)</label>
                    <Input
                        type="number"
                        id="price"
                        name="price"
                        value={newProduct.price}
                        onChange={handleInputChange}
                        disabled={isSubmitting}
                        step="0.01"
                        min="0.01"
                        placeholder="e.g., 599.99"
                        required
                    />
                </div>
                <Button
                    type="submit"
                    className="w-full"
                    disabled={isSubmitting}
                >
                    {isSubmitting ? (
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    ) : (
                        <PlusCircle className="w-4 h-4 mr-2" />
                    )}
                    {isSubmitting ? 'Adding...' : 'Save Product'}
                </Button>
            </form>
        </Card>
    );
};

// Component for an individual product item
const ProductItem = ({ product, onProductDeleted, onProductUpdated, onError }) => {
    const [isEditing, setIsEditing] = useState(false);
    const [editData, setEditData] = useState({ 
        name: product.name, 
        price: product.price ? product.price.toString() : '0.00' 
    });
    const [isSaving, setIsSaving] = useState(false);

    const handleEditChange = (e) => {
        const { name, value } = e.target;
        setEditData(prev => ({ ...prev, [name]: value }));
    };

    const handleSave = async () => {
        setIsSaving(true);
        onError(null);

        const dataToSend = {
            name: editData.name.trim(),
            price: parseFloat(editData.price), 
        };

        if (!dataToSend.name || isNaN(dataToSend.price) || dataToSend.price <= 0) {
            onError(`Invalid data for Product #${product.ID}. Please check name and price.`);
            setIsSaving(false);
            return;
        }

        try {
            await api.updateProduct(product.ID, dataToSend);
            onProductUpdated(product.ID, dataToSend);
            setIsEditing(false);
        } catch (e) {
            console.error("Error updating product:", e);
            onError("Could not update product. Check server logs.");
        } finally {
            setIsSaving(false);
        }
    };

    const handleDelete = async () => {
        onError(null);
        try {
            await api.deleteProduct(product.ID);
            onProductDeleted(product.ID);
        } catch (e) {
            console.error("Error deleting product:", e);
            onError("Could not delete product. Check server logs.");
        }
    };

    if (isEditing) {
        return (
            <li className="flex flex-col p-4 bg-indigo-50 rounded-lg shadow-inner border border-indigo-200 space-y-3">
                <div className="flex space-x-3 items-center">
                    <Input
                        type="text"
                        name="name"
                        value={editData.name}
                        onChange={handleEditChange}
                        disabled={isSaving}
                        className="flex-1 text-base font-medium"
                        placeholder="Product Name"
                    />
                    <Input
                        type="number"
                        name="price"
                        value={editData.price}
                        onChange={handleEditChange}
                        disabled={isSaving}
                        step="0.01"
                        className="w-24 text-base font-bold text-right"
                    />
                </div>
                <div className="flex justify-end space-x-2">
                    <Button
                        onClick={handleSave}
                        variant="default" // Using default for success in shadcn
                        size="sm"
                        disabled={isSaving}
                    >
                        {isSaving ? <Loader2 className="w-4 h-4 mr-1 animate-spin" /> : <Save className="w-4 h-4 mr-1" />}
                        Save Changes
                    </Button>
                    <Button
                        onClick={() => setIsEditing(false)}
                        variant="secondary"
                        size="sm"
                        className="p-2"
                        title="Cancel"
                    >
                        <X className="w-4 h-4" />
                    </Button>
                </div>
            </li>
        );
    }

    // Default view
    return (
        <li className="flex items-center justify-between p-4 bg-white rounded-lg border border-gray-200 hover:shadow-md transition duration-150">
            <div className="flex-1 min-w-0">
                <p className="text-lg font-semibold text-gray-900 truncate">{product.name}</p>
                <p className="text-xs text-gray-500 mt-1">
                    ID: {product.ID} | Updated: {new Date(product.UpdatedAt).toLocaleDateString()}
                </p>
            </div>
            <div className="flex items-center space-x-3">
                <span className="text-xl font-bold text-green-600">
                    ${product.price ? product.price.toFixed(2) : '0.00'}
                </span>
                <Button
                    onClick={() => setIsEditing(true)}
                    variant="ghost"
                    size="icon"
                    className="text-indigo-500 hover:text-indigo-700"
                    title="Edit Product"
                >
                    <Pencil className="w-5 h-5" />
                </Button>
                <Button
                    onClick={handleDelete}
                    variant="ghost"
                    size="icon"
                    className="text-red-500 hover:text-red-700"
                    title="Delete Product"
                >
                    <Trash2 className="w-5 h-5" />
                </Button>
            </div>
        </li>
    );
};


const ProductList = ({ products, loading, error, onProductDeleted, onProductUpdated, onError }) => {
    
    if (loading) {
        return (
            <div className="flex justify-center items-center p-10 text-indigo-500">
                <Loader2 className="w-8 h-8 animate-spin mr-3" />
                Loading inventory data...
            </div>
        );
    }
    
    // Sort by ID descending (newest first)
    const sortedProducts = [...products].sort((a, b) => b.ID - a.ID); 
    
    if (sortedProducts.length === 0 && !error) {
        return (
            <div className="text-center p-10 border border-dashed rounded-xl bg-gray-50">
                <p className="text-gray-500">
                    No products found. Use the form to the left to add your first item!
                </p>
            </div>
        );
    }

    return (
        <ul className="space-y-3">
            {sortedProducts.map((product) => (
                <ProductItem 
                    key={product.ID} 
                    product={product} 
                    onProductDeleted={onProductDeleted} 
                    onProductUpdated={onProductUpdated}
                    onError={onError}
                />
            ))}
        </ul>
    );
};


// --- Main App Component ---
const App = () => {
    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const loadProducts = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await api.fetchProducts();
            setProducts(data);
        } catch (e) {
            console.error("Failed to fetch products:", e);
            setError("Failed to load products from Go API. Please ensure the Go server is running on http://localhost:8080.");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadProducts();
    }, [loadProducts]);

    const handleProductAdded = () => {
        loadProducts();
    };

    const handleProductDeleted = (deletedId) => {
        // After deletion, refresh the list to ensure data is consistent
        loadProducts();
    };

    const handleProductUpdated = (updatedId, updatedData) => {
        // After update, refresh the list to ensure sorting (newest first) is respected
        loadProducts();
    };
    
    const handleSetError = (msg) => {
        setError(msg);
    };

    return (
        <div className="min-h-screen bg-gray-50 p-4 sm:p-8 font-sans antialiased">
            <div className="max-w-6xl mx-auto">
                <header className="py-8 mb-10 border-b border-indigo-200">
                    <h1 className="text-4xl font-extrabold text-indigo-700 tracking-tight">
                        Go/GORM & Next.js/Shadcn Inventory Manager
                    </h1>
                    <p className="text-gray-500 mt-2 text-lg">
                        A robust, full CRUD template: Go backend (GORM/SQLite) serves the API, and the React frontend provides the Shadcn-style UI.
                    </p>
                </header>

                {error && (
                    <div className="mb-6">
                        <Alert variant="destructive">
                            <AlertTriangle className="h-4 w-4" />
                            <AlertTitle>API Connection Error</AlertTitle>
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    </div>
                )}

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Product Creation Form */}
                    <div className="lg:col-span-1">
                        <ProductForm 
                            onProductAdded={handleProductAdded}
                            onError={handleSetError}
                        />
                    </div>

                    {/* Product List */}
                    <div className="lg:col-span-2">
                         <Card className="p-6">
                            <h2 className="text-2xl font-bold text-gray-800 mb-6">
                                Current Inventory ({products.length} Items)
                            </h2>
                            <ProductList 
                                products={products}
                                loading={loading}
                                error={error}
                                onProductDeleted={handleProductDeleted}
                                onProductUpdated={handleProductUpdated}
                                onError={handleSetError}
                            />
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default App;