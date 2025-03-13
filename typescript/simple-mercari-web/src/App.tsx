import { useState } from 'react';
import './App.css';
import { ItemList } from '~/components/ItemList';
import { Listing } from '~/components/Listing';
import { Search } from '~/components/Search';
import { Item, searchItem } from '~/api';

function App() {
  const [reload, setReload] = useState(true);
  const [searchResults, setSearchResults] = useState<Item[] | null>(null);

  const handleSearch = (results: Item[]) => {
    setSearchResults(results);
  };

  const handleCategoryClick = async (category: string) => {
    try {
      const data = await searchItem(category);
      setSearchResults(data.items);
    } catch (error) {
      console.error('Search error:', error);
      alert('Failed to search for items by category');
    }
  };

  return (
    <div>
      <header className="Title">
        <div>
          <p>
            <b>Simple Mercari</b>
          </p>
        </div>
        <div>
          <Search onSearchCompleted={handleSearch} />
        </div>

      </header>

      <div>
        <Listing onListingCompleted={() => setReload(true)} />
      </div>

      <div>
        <ItemList
          items={searchResults || undefined}
          reload={!searchResults && reload}
          onLoadCompleted={() => setReload(false)}
          onCategoryClick={handleCategoryClick}
        />
      </div>
    </div>
  );
}

export default App;
