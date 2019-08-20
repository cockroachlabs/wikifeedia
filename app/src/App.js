import React from 'react';
// import logo from './logo.svg';
import './App.css';

import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';
import gql from 'graphql-tag';
import { useQuery, ApolloProvider as ApolloHooksProvider } from '@apollo/react-hooks';
import { ApolloProvider } from 'react-apollo';

const client = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: "/graphqlhttp" }),
  cache: new InMemoryCache(),
});

const GET_FEED = gql`
query Feed($project: String!, $offset: Int, $limit: Int) {
  articles(project: $project, offset: $offset, limit: $limit) {
    project
    abstract
    article
    articleURL
    dailyViews
    imageURL
    thumbnailURL
    title
  }
}`;


class Feed extends React.Component {
  componentDidMount() {
    window.addEventListener("scroll", this.handleOnScroll);
  }

  componentWillUnmount() {
    window.removeEventListener("scroll", this.handleOnScroll);
  }

  handleOnScroll = () => {
    // http://stackoverflow.com/questions/9439725/javascript-how-to-detect-if-browser-window-is-scrolled-to-bottom
    var scrollTop =
      (document.documentElement && document.documentElement.scrollTop) ||
      document.body.scrollTop;
    var scrollHeight =
      (document.documentElement && document.documentElement.scrollHeight) ||
      document.body.scrollHeight;
    var clientHeight =
      document.documentElement.clientHeight || window.innerHeight;
    var scrolledToBottom = Math.ceil(scrollTop + clientHeight) >= scrollHeight;
    if (scrolledToBottom) {
      this.props.onLoadMore();
    }
  };

  render() {
    if (this.props.error) return `Error! ${this.props.error.message}`;
    if (!this.props.articles && this.props.loading) return <p>Loading....</p>;
    const articles = this.props.articles || [];
    return (
      <div className="Articles">
        {articles.map(({
          project,
          article,
          abstract,
          articleURL,
          dailyViews,
          imageURL,
          thumbnailURL,
          title
        }, idx) => (
        <div className="Article" id={article} key={idx} project={project}>
          <div className="ArticleImageContainer">
            <div className="ArticleImage">
              <a href={articleURL} target="_blank" rel="noopener noreferrer">
                <img src={thumbnailURL} alt={title}/>
              </a>
            </div>
          </div>
          <div className="ArticleContent">
            <h2 className="ArticleTitle">{title}</h2>
            <p className="ArticleAbstractText">
              {abstract}
            </p>
          </div>
        </div>)
      )}
      </div>
    );
  }
}

function FeedApp({ project }) {
  const {loading, error, data, fetchMore } = useQuery(GET_FEED, {
    variables: {
      project: project,
      offset: 0,
      limit: 10
    },
    fetchPolicy: "cache-and-network"
  });
  return (
    <Feed
      key="feed"
      articles={data.articles || []}
      loading={loading}
      error={error}
      onLoadMore={() =>
        fetchMore({
          variables: {
            offset: data.articles.length
          },
          updateQuery: (prev, { fetchMoreResult }) => {
            if (!fetchMoreResult) return prev;
            return Object.assign({}, prev, {
              articles: [...prev.articles, ...fetchMoreResult.articles]
            });
          }
        })
      }
    />
  )
}

const languages = [
  "en",
	"fr",
	"es",
	"de",
	"ru",
	"ja",
	"nl",
	"it",
	"sv",
	"pl",
	"vi",
	"pt",
	"ar",
	"zh",
  "uk",
  "ro",
  "bg",
  "th"
]



class App extends React.Component {
  state = {
    project: "en",
  }
  setProject(project) {
    this.setState({project: project});
  }
  render()  {
    const langTabs = languages.map((lang) => {
      const onClick = () => {this.setProject(lang)};
      return (<div onClick={onClick} key={lang}>{lang}</div>)
    });
    function App({project}) {
      return (
      <div className="App">
        <div className="LangTabs">
          {langTabs}
        </div>
        <div className="AppHeaderContainer">
          <div className="App-header">
            <h1 className="AppTitle">Wikifeedia</h1> 
          </div>
        </div>
        
        <FeedApp project={project}/>
      </div>
      );
    };
    return (
      <ApolloProvider client={client}>
        <ApolloHooksProvider client={client}>
          <App project={this.state.project}/>
        </ApolloHooksProvider>
      </ApolloProvider>
    );
  }
}

export default App;

