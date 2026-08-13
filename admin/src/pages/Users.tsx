import { useState, useEffect } from 'react';
import { saaskitClient } from '../services/saaskit';
import { useAuth } from '../context/AuthContext';

interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  status: string;
  created_at: string;
  metadata_public?: Record<string, unknown>;
  metadata_private?: Record<string, unknown>;
}

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showMetadataModal, setShowMetadataModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [newEmail, setNewEmail] = useState('');
  const [newName, setNewName] = useState('');
  const [metadataPublic, setMetadataPublic] = useState('');
  const [metadataPrivate, setMetadataPrivate] = useState('');
  const { accessToken } = useAuth();

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const fetchedUsers = await saaskitClient.users.listAllUsers(accessToken);
      setUsers(fetchedUsers);
    } catch (error) {
      console.error('Failed to fetch users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newEmail) return;
    try {
      await saaskitClient.users.createUserAdmin(
        accessToken,
        {
          email: newEmail,
          first_name: newName || undefined
        }
      );
      await fetchUsers();
      setNewEmail('');
      setNewName('');
      setShowAddModal(false);
    } catch (error) {
      console.error('Failed to create user:', error);
      alert('Failed to create user. Please try again.');
    }
  };

  const toggleUserStatus = async (id: string) => {
    try {
      const user = users.find(u => u.id === id);
      if (!user) return;
      const newStatus = user.status === 'active' ? 'disabled' : 'active';
      await saaskitClient.users.updateUserStatus(accessToken, id, newStatus);
      await fetchUsers();
    } catch (error) {
      console.error('Failed to update user status:', error);
    }
  };

  const handleEditMetadata = (user: User) => {
    setSelectedUser(user);
    setMetadataPublic(user.metadata_public ? JSON.stringify(user.metadata_public, null, 2) : '');
    setMetadataPrivate(user.metadata_private ? JSON.stringify(user.metadata_private, null, 2) : '');
    setShowMetadataModal(true);
  };

  const handleSaveMetadata = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUser) return;

    let parsedPublic: Record<string, unknown> | undefined;
    let parsedPrivate: Record<string, unknown> | undefined;

    try {
      if (metadataPublic.trim()) {
        parsedPublic = JSON.parse(metadataPublic);
      }
      if (metadataPrivate.trim()) {
        parsedPrivate = JSON.parse(metadataPrivate);
      }

      await saaskitClient.users.updateUserMetadataAdmin(
        accessToken,
        selectedUser.id,
        {
          metadata_public: parsedPublic,
          metadata_private: parsedPrivate
        }
      );

      await fetchUsers();
      setShowMetadataModal(false);
      setSelectedUser(null);
      setMetadataPublic('');
      setMetadataPrivate('');
    } catch (error) {
      if (error instanceof SyntaxError) {
        alert('Invalid JSON format. Please check your input.');
      } else {
        console.error('Failed to update user metadata:', error);
        alert('Failed to update metadata. Please try again.');
      }
    }
  };

  const filteredUsers = users.filter(
    (user) =>
      user.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (user.first_name && user.first_name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (user.last_name && user.last_name.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  return (
    <div className="animate-fade-in">
      <div className="sm:flex sm:items-center sm:justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Users</h1>
          <p className="mt-2 text-gray-600">
            Manage user accounts and their access to the system.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            type="button"
            onClick={() => setShowAddModal(true)}
            className="inline-flex items-center px-4 py-2.5 border border-transparent rounded-xl text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-all duration-200 shadow-sm cursor-pointer"
          >
            <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 4v16m8-8H4" />
            </svg>
            Add User
          </button>
        </div>
      </div>

      <div className="mb-6">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <input
            type="text"
            placeholder="Search users by email or name..."
            className="block w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 bg-white shadow-xs"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-100">
          <thead className="bg-gray-50/50">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                User
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Created
              </th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {loading ? (
              <tr>
                <td colSpan={4} className="px-6 py-16 text-center text-sm text-gray-500">
                  <div className="flex items-center justify-center space-x-3">
                    <svg className="animate-spin h-5 w-5 text-primary-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="font-medium text-gray-600">Loading users...</span>
                  </div>
                </td>
              </tr>
            ) : filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-16 text-center text-sm text-gray-400">
                  <svg className="mx-auto h-12 w-12 text-gray-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  <p className="font-medium">No users found</p>
                </td>
              </tr>
            ) : (
              filteredUsers.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50/55 transition-colors duration-150">
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-primary-50 border border-primary-100 flex items-center justify-center">
                          <span className="text-sm font-semibold text-primary-600">
                            {user.email.charAt(0).toUpperCase()}
                          </span>
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-semibold text-gray-900">{user.email}</div>
                        <div className="text-sm text-gray-500">{user.first_name && user.last_name ? `${user.first_name} ${user.last_name}` : user.first_name || user.last_name || 'No name'}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold border ${
                      user.status === 'active' 
                        ? 'bg-green-50 text-green-700 border-green-200' 
                        : 'bg-amber-50 text-amber-700 border-amber-200'
                    }`}>
                      <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
                        user.status === 'active' ? 'bg-green-500' : 'bg-amber-500'
                      }`}></span>
                      {user.status}
                    </span>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-medium">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-right text-sm font-medium">
                    <button 
                      onClick={() => handleEditMetadata(user)}
                      className="text-xs font-semibold px-3 py-1.5 rounded-lg border border-primary-200 text-primary-600 hover:bg-primary-50 transition-all duration-200 hover:shadow-xs cursor-pointer mr-2"
                    >
                      Edit Metadata
                    </button>
                    <button 
                      onClick={() => toggleUserStatus(user.id)}
                      className={`text-xs font-semibold px-3 py-1.5 rounded-lg border transition-all duration-200 hover:shadow-xs cursor-pointer ${
                        user.status === 'active'
                          ? 'border-amber-200 text-amber-600 hover:bg-amber-50'
                          : 'border-green-200 text-green-600 hover:bg-green-50'
                      }`}
                    >
                      {user.status === 'active' ? 'Disable' : 'Enable'}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="fixed inset-0 bg-gray-900/40 backdrop-blur-xs flex items-center justify-center z-50 animate-fade-in">
          <div className="bg-white rounded-2xl shadow-xl max-w-md w-full mx-4 border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between">
              <h3 className="text-lg font-bold text-gray-900">Add New User</h3>
              <button 
                onClick={() => setShowAddModal(false)}
                className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-55"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleAddUser}>
              <div className="p-6 space-y-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Email Address</label>
                  <input
                    type="email"
                    required
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="user@example.com"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Full Name (Optional)</label>
                  <input
                    type="text"
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="Jane Doe"
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                  />
                </div>
              </div>
              <div className="px-6 py-4.5 bg-gray-50 border-t border-gray-100 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="px-4 py-2 border border-gray-200 rounded-xl text-sm font-semibold text-gray-600 bg-white hover:bg-gray-50 transition-all duration-200 cursor-pointer"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-xl text-sm font-semibold hover:bg-primary-700 transition-all duration-200 shadow-sm cursor-pointer"
                >
                  Add User
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showMetadataModal && selectedUser && (
        <div className="fixed inset-0 bg-gray-900/40 backdrop-blur-xs flex items-center justify-center z-50 animate-fade-in">
          <div className="bg-white rounded-2xl shadow-xl max-w-2xl w-full mx-4 border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between">
              <div>
                <h3 className="text-lg font-bold text-gray-900">Edit User Metadata</h3>
                <p className="text-sm text-gray-500 mt-1">{selectedUser.email}</p>
              </div>
              <button 
                onClick={() => {
                  setShowMetadataModal(false);
                  setSelectedUser(null);
                  setMetadataPublic('');
                  setMetadataPrivate('');
                }}
                className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-55"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleSaveMetadata}>
              <div className="p-6 space-y-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">
                    Public Metadata
                    <span className="text-gray-400 font-normal ml-2">(visible to client applications)</span>
                  </label>
                  <textarea
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm font-mono min-h-32"
                    placeholder='{"key": "value"}'
                    value={metadataPublic}
                    onChange={(e) => setMetadataPublic(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">
                    Private Metadata
                    <span className="text-gray-400 font-normal ml-2">(server-side only)</span>
                  </label>
                  <textarea
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm font-mono min-h-32"
                    placeholder='{"key": "value"}'
                    value={metadataPrivate}
                    onChange={(e) => setMetadataPrivate(e.target.value)}
                  />
                </div>
              </div>
              <div className="px-6 py-4.5 bg-gray-50 border-t border-gray-100 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowMetadataModal(false);
                    setSelectedUser(null);
                    setMetadataPublic('');
                    setMetadataPrivate('');
                  }}
                  className="px-4 py-2 border border-gray-200 rounded-xl text-sm font-semibold text-gray-600 bg-white hover:bg-gray-50 transition-all duration-200 cursor-pointer"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-xl text-sm font-semibold hover:bg-primary-700 transition-all duration-200 shadow-sm cursor-pointer"
                >
                  Save Metadata
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
